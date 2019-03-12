package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/apex/log"

	"github.com/hink/SquadSheets/pkg/models"

	"google.golang.org/api/sheets/v4"
)

const NEWLINE = "\r\n"
const ASCWhitelistURL = "https://docs.google.com/document/d/1tMzGySrFdXIKW1JG9sKTJiEvsEGDM0oT4tzuxojXWpc/export?format=txt"

func WriteAdminsFile(srv *sheets.Service, sheetID, configDir string, includeASCWhitelist bool) error {
	adminRoles := make(map[string]models.AdminRole)
	admins := []models.User{}
	whitelist := []models.User{}

	// get admin roles
	dataRange := fmt.Sprintf("%s!A2:B", cfg.Sheets.SheetAdminRoles)

	resp, err := srv.Spreadsheets.Values.Get(sheetID, dataRange).Do()
	if err != nil {
		err = fmt.Errorf("unable to retrieve data from admin roles sheet %v", err)
		return err
	}
	if len(resp.Values) > 0 {
		for _, row := range resp.Values {
			r := rowToAdminRole(row)
			adminRoles[r.Name] = r
		}
	}

	// check for whitelist role and add it if necessary
	hasAdminRole := false
	for role, _ := range adminRoles {
		if role == "Whitelist" {
			hasAdminRole = true
			break
		}
	}
	if !hasAdminRole {
		r := models.AdminRole{
			Name:  "Whitelist",
			Value: "reserve",
		}
		adminRoles["Whitelist"] = r
	}

	// get admins
	for _, s := range cfg.Sheets.SheetsAdmin {
		dataRange = fmt.Sprintf("%s!A2:D", s)

		resp, err = srv.Spreadsheets.Values.Get(sheetID, dataRange).Do()
		if err != nil {
			err = fmt.Errorf("Unable to retrieve data from admin %s sheet. %v", s, err)
			return err
		}

		if len(resp.Values) > 0 {
			for _, row := range resp.Values {
				u := rowToUser(row)
				admins = append(admins, u)
			}
		}
	}

	// get whitelist
	for _, s := range cfg.Sheets.SheetsWhitelist {
		dataRange = fmt.Sprintf("%s!A2:C", s)

		resp, err = srv.Spreadsheets.Values.Get(sheetID, dataRange).Do()
		if err != nil {
			err = fmt.Errorf("Unable to retrieve data from whitelist %s sheet. %v", s, err)
			return err
		}

		if len(resp.Values) > 0 {
			for _, row := range resp.Values {
				u := rowToWhitelistUser(row)
				u.Role = adminRoles["Whitelist"].Name
				whitelist = append(whitelist, u)
			}
		}
	}

	// write file
	timeGenerated := time.Now().Format("Mon Jan _2 15:04:05 2006")
	fileData := "// THIS FILE SHOULD NOT BE MODIFIED MANUALLY. IT IS MANAGED VIA A SCHEDULED TASK AND GOOGLE SHEETS" + NEWLINE
	fileData = "// last generated " + timeGenerated

	// Admin roles
	fileData += NEWLINE + NEWLINE + "// ADMIN ROLES --------------" + NEWLINE
	for _, xRole := range adminRoles {
		fileData += fmt.Sprintf("Group=%s:%s%s", xRole.Name, xRole.Value, NEWLINE)
	}

	// VKNG Admins
	fileData += NEWLINE + NEWLINE + "// ADMINS --------------" + NEWLINE
	for _, admin := range admins {
		if admin.Steam64 == "" {
			continue // we dont have their 64 anyway
		}
		fileData += fmt.Sprintf("Admin=%s:%s\t\t//%s", admin.Steam64, admin.Role, admin.Name)
		if admin.Notes != "" {
			fileData += fmt.Sprintf(" - %s", admin.Notes)
		}
		fileData += NEWLINE
	}

	// Whitelist
	fileData += NEWLINE + NEWLINE + "// WHITELIST --------------" + NEWLINE
	for _, user := range whitelist {
		if user.Steam64 == "" {
			continue // we dont have their 64 anyway
		}
		fileData += fmt.Sprintf("Admin=%s:%s\t\t//%s", user.Steam64, user.Role, user.Name)
		if user.Notes != "" {
			fileData += fmt.Sprintf(" - %s", user.Notes)
		}
		fileData += NEWLINE
	}

	// ASC
	if includeASCWhitelist {
		response, err := http.Get(ASCWhitelistURL)
		if err != nil {
			log.Errorf("Error downloading ASC Whitelist: %s", err.Error())
		} else {
			ascData, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Errorf("Error reading ASC Whitelist: %s", err.Error())
			} else {
				fileData += NEWLINE + NEWLINE + string(ascData)
			}
		}
		defer response.Body.Close()
	}

	// write file
	filePath := filepath.Join(configDir, "Admins.cfg")
	err = ioutil.WriteFile(filePath, []byte(fileData), 0644)
	if err != nil {
		return err
	}

	return nil
}

func rowToAdminRole(row []interface{}) models.AdminRole {
	return models.AdminRole{
		Name:  row[0].(string),
		Value: row[1].(string),
	}
}

func rowToUser(row []interface{}) models.User {
	// true up weird workaround
	if len(row) < 4 {
		row = append(row, "")
	}
	return models.User{
		Name:    row[0].(string),
		Steam64: row[1].(string),
		Role:    row[2].(string),
		Notes:   row[3].(string),
	}
}

func rowToWhitelistUser(row []interface{}) models.User {
	// true up weird workaround
	if len(row) < 4 {
		row = append(row, "")
	}
	return models.User{
		Name:    row[0].(string),
		Steam64: row[1].(string),
		Notes:   row[2].(string),
	}
}
