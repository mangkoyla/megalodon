package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/FoolVPN-ID/Megalodon/common/helper"
	logger "github.com/FoolVPN-ID/Megalodon/log"
	"github.com/FoolVPN-ID/Megalodon/sandbox"
	"github.com/FoolVPN-ID/Megalodon/telegram/bot"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type databaseStruct struct {
	client          *sql.DB
	dbUrl           string
	dbToken         string
	logger          *logger.LoggerStruct
	rawAccountTotal int
	uniqueIds       []string
	queries         []string
	ErrorValues     []string
}

func MakeDatabase() *databaseStruct {
	dbInstance := databaseStruct{
		dbUrl:   os.Getenv("TURSO_DATABASE_URL"),
		dbToken: os.Getenv("TURSO_AUTH_TOKEN"),
		logger:  logger.MakeLogger(),
	}

	dbInstance.connect()

	return &dbInstance
}

func (db *databaseStruct) connect() {
	url := fmt.Sprintf("%s?authToken=%s", db.dbUrl, db.dbToken)
	client, err := sql.Open("libsql", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", url, err)
		os.Exit(1)
	}

	db.client = client
}

func (db *databaseStruct) Close() {
	db.logger.Info("Closing client...")
	if err := db.client.Close(); err != nil {
		db.logger.Error(err.Error())
	}
}

func (db *databaseStruct) createTableSafe() {
	var (
		crateTableQuery = `CREATE TABLE IF NOT EXISTS proxies (
			ID INTEGER PRIMARY KEY,
			SERVER TEXT,
			IP TEXT,
			SERVER_PORT INT8,
			UUID TEXT,
			PASSWORD TEXT,
			SECURITY TEXT,
			ALTER_ID INT2,
			METHOD TEXT,
			PLUGIN TEXT,
			PLUGIN_OPTS TEXT,
			HOST TEXT,
			TLS INT2,
			TRANSPORT TEXT,
			PATH TEXT,
			SERVICE_NAME TEXT,
			INSECURE INT2,
			SNI TEXT,
			REMARK TEXT,
			CONN_MODE TEXT,
			COUNTRY_CODE TEXT,
			REGION TEXT,
			ORG TEXT,
			VPN TEXT
		);`
	)

	if _, err := db.client.Query(crateTableQuery); err == nil {
		db.logger.Info("Database successfully created!")
	} else {
		db.logger.Error(err.Error())
		os.Exit(1)
	}
}

func (db *databaseStruct) Save(results []sandbox.TestResultStruct) error {
	db.createTableSafe()
	db.queries = append(db.queries, "DELETE FROM proxies;")
	db.queries = append(db.queries, db.buildInsertQuery(results)...)

	tgb := bot.MakeTGgBot()
	tgb.SendTextFileToAdmin(fmt.Sprintf("query_%v.txt", time.Now().Unix()), strings.Join(db.queries, "\n"), "DB Query")
	if len(db.ErrorValues) > 0 {
		tgb.SendTextFileToAdmin(fmt.Sprintf("error_%v.txt", time.Now().Unix()), strings.Join(db.ErrorValues, "\n"), "Error Values")
	}

	// Begin transaction
	transaction, err := db.client.Begin()
	if err != nil {
		db.logger.Error(err.Error())
		return err
	}

	for _, dbQuery := range db.queries {
		if _, err := transaction.Exec(dbQuery); err != nil {
			transaction.Rollback()
			db.logger.Error(err.Error())
			return err
		}
	}

	if err := transaction.Commit(); err != nil {
		transaction.Rollback()
		db.logger.Error(err.Error())
		return err
	} else {
		db.logger.Info("=========================")
		db.logger.Success("Data successfully saved!")
		db.logger.Info(fmt.Sprintf("Total raw account: %d", db.rawAccountTotal))
		db.logger.Info(fmt.Sprintf("Total account saved: %d", len(db.uniqueIds)))

		// Report
		tgb.SendTextToAdmin(fmt.Sprintf("Account saved: %d", len(db.uniqueIds)))
	}

	return nil
}

func (db *databaseStruct) buildInsertQuery(results []sandbox.TestResultStruct) []string {
	db.rawAccountTotal = len(results)

	tableFieldValues := []DatabaseFieldStruct{}
	for _, result := range results {
		var (
			fieldValues = DatabaseFieldStruct{}
			outbound    = result.Outbound
		)

		var (
			outboundMapping = map[string]any{}
			outboundAny, _  = outbound.RawOptions()
			outboundByte, _ = json.Marshal(outboundAny)
		)
		json.Unmarshal(outboundByte, &outboundMapping)

		// Geoip
		fieldValues.Ip = result.ConfigGeoip.IP
		fieldValues.CountryCode = result.ConfigGeoip.Country
		fieldValues.Region = helper.GetRegionFromCC(fieldValues.CountryCode)
		fieldValues.Org = result.ConfigGeoip.AsOrganization

		// Common
		fieldValues.VPN = outbound.Type
		fieldValues.Server = outboundMapping["server"].(string)
		fieldValues.ServerPort = int(outboundMapping["server_port"].(float64))

		// Here we go assertion hell
		if uuid, ok := outboundMapping["uuid"].(string); ok {
			fieldValues.UUID = uuid
		}
		if password, ok := outboundMapping["password"].(string); ok {
			fieldValues.Password = password
		}
		if security, ok := outboundMapping["security"].(string); ok {
			fieldValues.Security = security
		}
		if alterId, ok := outboundMapping["alter_id"].(int); ok {
			fieldValues.AlterId = alterId
		}
		if method, ok := outboundMapping["method"].(string); ok {
			fieldValues.Method = method
		}
		if plugin, ok := outboundMapping["plugin"].(string); ok {
			fieldValues.Plugin = plugin
		}
		if pluginOpts, ok := outboundMapping["plugin_opts"].(string); ok {
			fieldValues.PluginOpts = pluginOpts
		}

		// Transport
		if outboundMapping["transport"] != nil {
			transportMapping := outboundMapping["transport"].(map[string]any)
			if transportType, ok := transportMapping["type"].(string); ok {
				fieldValues.Transport = transportType
			}
			if serviceName, ok := transportMapping["service_name"].(string); ok {
				fieldValues.ServiceName = serviceName
			}
			if path, ok := transportMapping["path"].(string); ok {
				fieldValues.Path = path
			}
			if host, ok := transportMapping["host"].(string); ok {
				fieldValues.Host = host
			}
			if transportMapping["headers"] != nil {
				headersMapping := transportMapping["headers"].(map[string]any)
				if host, ok := headersMapping["Host"].(string); ok {
					fieldValues.Host = host
				}
			}
		}

		// TLS
		tlsStr := "NTLS"
		if outboundMapping["tls"] != nil {
			tlsMapping := outboundMapping["tls"].(map[string]any)
			if enabled, ok := tlsMapping["enabled"].(bool); ok {
				fieldValues.TLS = enabled
				if enabled {
					tlsStr = "TLS"
				}
			}
			if insecure, ok := tlsMapping["insecure"].(bool); ok {
				fieldValues.Insecure = insecure
			}
			if sni, ok := tlsMapping["server_name"].(string); ok {
				fieldValues.SNI = sni
			}
		}

		for _, connMode := range result.TestPassed {
			fieldValues.ConnMode = connMode
			fieldValues.Remark = strings.ToUpper(fmt.Sprintf("%d %s %s %s %s %s", len(tableFieldValues)+1, helper.CCToEmoji(fieldValues.CountryCode), fieldValues.Org, fieldValues.Transport, connMode, tlsStr))

			// Check if same account exists
			if !db.checkIsExists(fieldValues) {
				tableFieldValues = append(tableFieldValues, fieldValues)
			}
		}
	}

	// Build queries
	values := []string{}
	for _, fieldValue := range tableFieldValues {
		value := "("

		value += fmt.Sprintf("'%s', ", fieldValue.Server)
		value += fmt.Sprintf("'%s', ", fieldValue.Ip)
		value += fmt.Sprintf("%d, ", fieldValue.ServerPort)
		value += fmt.Sprintf("'%s', ", fieldValue.UUID)
		value += fmt.Sprintf("'%s', ", fieldValue.Password)
		value += fmt.Sprintf("'%s', ", fieldValue.Security)
		value += fmt.Sprintf("%d, ", fieldValue.AlterId)
		value += fmt.Sprintf("'%v', ", fieldValue.Method)
		value += fmt.Sprintf("'%v', ", fieldValue.Plugin)
		value += fmt.Sprintf("'%v', ", fieldValue.PluginOpts)
		value += fmt.Sprintf("'%v', ", fieldValue.Host)
		value += fmt.Sprintf("%t, ", fieldValue.TLS)
		value += fmt.Sprintf("'%s', ", fieldValue.Transport)
		value += fmt.Sprintf("'%s', ", fieldValue.Path)
		value += fmt.Sprintf("'%s', ", fieldValue.ServiceName)
		value += fmt.Sprintf("%t, ", fieldValue.Insecure)
		value += fmt.Sprintf("'%s', ", fieldValue.SNI)
		value += fmt.Sprintf("'%s', ", fieldValue.Remark)
		value += fmt.Sprintf("'%s', ", fieldValue.ConnMode)
		value += fmt.Sprintf("'%s', ", fieldValue.CountryCode)
		value += fmt.Sprintf("'%s', ", fieldValue.Region)
		value += fmt.Sprintf("'%s', ", fieldValue.Org)
		value += fmt.Sprintf("'%s'", fieldValue.VPN)

		value += ")"

		values = append(values, value)
	}

	baseInsertQuery := `INSERT INTO proxies (
		SERVER,
		IP,
		SERVER_PORT,
		UUID, PASSWORD,
		SECURITY,
		ALTER_ID,
		METHOD,
		PLUGIN,
		PLUGIN_OPTS,
		HOST,
		TLS,
		TRANSPORT,
		PATH,
		SERVICE_NAME,
		INSECURE,
		SNI,
		REMARK,
		CONN_MODE,
		COUNTRY_CODE,
		REGION,
		ORG,
		VPN
	) VALUES`

	// Filter bad and build insert queries
	insertQueries := []string{}

	for i := 0; i < len(values); i += 500 {
		end := i + 500
		if end > len(values) {
			end = len(values)
		}
		insertQueries = append(insertQueries, fmt.Sprintf(`%s %s;`, baseInsertQuery, strings.Join(values[i:end], ",")))
	}

	return insertQueries
}

func (db *databaseStruct) makeUniqueId(field DatabaseFieldStruct) string {
	// Server Port, UUID, Password, Plugin Opts, Path, Transport, Conn Mode, Country, Org, VPN
	uid := fmt.Sprintf("%d_%s_%s_%s_%s_%s_%s_%s_%s_%s", field.ServerPort, field.UUID, field.Password, field.PluginOpts, field.Path, field.Transport, field.ConnMode, field.CountryCode, field.Org, field.VPN)

	return uid
}

func (db *databaseStruct) checkIsExists(field DatabaseFieldStruct) bool {
	uid := db.makeUniqueId(field)
	for _, existedUid := range db.uniqueIds {
		if existedUid == uid {
			return true
		}
	}

	db.uniqueIds = append(db.uniqueIds, uid)
	return false
}
