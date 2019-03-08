package main

import (
	"io/ioutil"

	"google.golang.org/api/sheets/v4"

	"fmt"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
)

func getSheetsService(secretPath string) (srv *sheets.Service, err error) {
	// Create GAPI context
	ctx := context.Background()

	b, err := ioutil.ReadFile(secretPath)
	if err != nil {
		err = fmt.Errorf("Unable to read Google client secret file: %v", err)
		return
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/sheets.googleapis.com-go-quickstart.json
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		err = fmt.Errorf("Unable to parse sheets client secret file to config: %v", err)
		return
	}
	client := getClient(ctx, config)

	srv, err = sheets.New(client)
	if err != nil {
		err = fmt.Errorf("Unable to retrieve Sheets Client %v", err)
		return
	}

	return
}

func getSheetProperties(srv *sheets.Service, sheetID string) (sheetsProperties []*sheets.Sheet, err error) {
	propsCall := srv.Spreadsheets.Get(sheetID)
	propsCall.Fields("sheets.properties")
	respProps, err := propsCall.Do()
	if err != nil {
		err = fmt.Errorf("Unable to get sheet properties: %v", err)
		return
	}
	sheetsProperties = respProps.Sheets
	return
}

func resizeSheet(srv *sheets.Service, rows, cols int64, sheetID string, props *sheets.SheetProperties) (err error) {
	deltaRows := int64(0)
	deltaCols := int64(0)

	if rows != -1 {
		deltaRows = rows - props.GridProperties.RowCount
	}
	if cols != -1 {
		deltaCols = cols - props.GridProperties.ColumnCount
	}

	if deltaRows == 0 && deltaCols == 0 {
		return
	}

	// build requests
	batchReq := new(sheets.BatchUpdateSpreadsheetRequest)

	if deltaRows != 0 {
		if deltaRows > 0 {
			// APPEND
			rowReq := new(sheets.AppendDimensionRequest)
			rowReq.Dimension = "ROWS"
			rowReq.Length = deltaRows
			rowReq.SheetId = props.SheetId
			batchReq.Requests = append(batchReq.Requests, &sheets.Request{AppendDimension: rowReq})
		} else {
			// DELETE
			rowReq := new(sheets.DeleteDimensionRequest)
			rowReq.Range = new(sheets.DimensionRange)
			rowReq.Range.Dimension = "ROWS"
			rowReq.Range.SheetId = props.SheetId
			rowReq.Range.StartIndex = 1
			rowReq.Range.EndIndex = deltaRows*-1 + 1
			batchReq.Requests = append(batchReq.Requests, &sheets.Request{DeleteDimension: rowReq})
		}
	}

	if deltaCols != 0 {
		if deltaCols > 0 {
			// APPEND
			colReq := new(sheets.AppendDimensionRequest)
			colReq.Dimension = "COLUMNS"
			colReq.Length = deltaCols
			colReq.SheetId = props.SheetId
			batchReq.Requests = append(batchReq.Requests, &sheets.Request{AppendDimension: colReq})
		} else {
			colReq := new(sheets.DeleteDimensionRequest)
			colReq.Range = new(sheets.DimensionRange)
			colReq.Range.Dimension = "COLUMNS"
			colReq.Range.SheetId = props.SheetId
			colReq.Range.StartIndex = 0
			colReq.Range.EndIndex = deltaCols * -1
			batchReq.Requests = append(batchReq.Requests, &sheets.Request{DeleteDimension: colReq})
		}
	}

	_, err = srv.Spreadsheets.BatchUpdate(sheetID, batchReq).Do()
	if err != nil {
		err = fmt.Errorf("Unable to update sheet dimensions: %v", err)
		return
	}

	return
}
