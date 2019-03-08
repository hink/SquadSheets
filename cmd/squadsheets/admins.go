package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/hink/SquadAdminSheets/pkg/models"

	"google.golang.org/api/sheets/v4"
)

const NEWLINE = "\r\n"

func WriteAdminsFile(srv *sheets.Service, sheetID, configDir string) error {
	adminRoles := make(map[string]models.AdminRole)
	adminsVkng := []models.User{}
	adminsOther := []models.User{}
	whitelist := []models.User{}

	// get admin roles
	dataRange := "admin_roles!A2:B"

	resp, err := srv.Spreadsheets.Values.Get(sheetID, dataRange).Do()
	if err != nil {
		err = fmt.Errorf("Unable to retrieve data from admin roles sheet. %v", err)
		return err
	}

	if len(resp.Values) > 0 {
		for _, row := range resp.Values {
			r := rowToAdminRole(row)
			adminRoles[r.Name] = r
		}
	}

	// get vkng admins
	dataRange = "admin_vkng!A2:D"

	resp, err = srv.Spreadsheets.Values.Get(sheetID, dataRange).Do()
	if err != nil {
		err = fmt.Errorf("Unable to retrieve data from admin vkng sheet. %v", err)
		return err
	}

	if len(resp.Values) > 0 {
		for _, row := range resp.Values {
			u := rowToUser(row)
			adminsVkng = append(adminsVkng, u)
		}
	}

	// get other admins
	dataRange = "admin_valhalla!A2:D"

	resp, err = srv.Spreadsheets.Values.Get(sheetID, dataRange).Do()
	if err != nil {
		err = fmt.Errorf("Unable to retrieve data from admin valhalla sheet. %v", err)
		return err
	}

	if len(resp.Values) > 0 {
		for _, row := range resp.Values {
			u := rowToUser(row)
			adminsOther = append(adminsOther, u)
		}
	}

	// get whitelist
	dataRange = "whitelist!A2:D"

	resp, err = srv.Spreadsheets.Values.Get(sheetID, dataRange).Do()
	if err != nil {
		err = fmt.Errorf("Unable to retrieve data from whitelist sheet. %v", err)
		return err
	}

	if len(resp.Values) > 0 {
		for _, row := range resp.Values {
			u := rowToUser(row)
			u.Role = adminRoles["Whitelist"].Name
			whitelist = append(whitelist, u)
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
	fileData += NEWLINE + NEWLINE + "// VKNG ADMINS --------------" + NEWLINE
	for _, xVkngAdmin := range adminsVkng {
		fileData += fmt.Sprintf("Admin=%s:%s //%s - %s%s", xVkngAdmin.Steam64, xVkngAdmin.Role, xVkngAdmin.Name, xVkngAdmin.Notes, NEWLINE)
	}

	// Community Admins
	fileData += NEWLINE + NEWLINE + "// COMMUNITY ADMINS --------------" + NEWLINE
	for _, xOtherAdmin := range adminsOther {
		fileData += fmt.Sprintf("Admin=%s:%s //%s - %s%s", xOtherAdmin.Steam64, xOtherAdmin.Role, xOtherAdmin.Name, xOtherAdmin.Notes, NEWLINE)
	}

	// Whitelist
	fileData += NEWLINE + NEWLINE + "// WHITELIST --------------" + NEWLINE
	for _, xWhitelist := range adminsOther {
		fileData += fmt.Sprintf("Admin=%s:%s //%s - %s%s", xWhitelist.Steam64, xWhitelist.Role, xWhitelist.Name, xWhitelist.Notes, NEWLINE)
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
	return models.User{
		Name:    row[0].(string),
		Steam64: row[1].(string),
		Role:    row[2].(string),
		Notes:   row[3].(string),
	}
}
