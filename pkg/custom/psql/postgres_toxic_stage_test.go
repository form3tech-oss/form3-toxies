package psql

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	toxiclient "github.com/Shopify/toxiproxy/v2/client"
	_ "github.com/lib/pq"
)

type PSQLTestStage struct {
	t         *testing.T
	db        *sql.DB
	psqlPort  int
	proxyPort int
}

type options struct {
	SearchText   string
	FailureType  FailureType
	FailAfter    int
	RecoverAfter int
}

const toxicName = "psql"
const proxyName = "postgres"
const ErrMessageFailure = "invalid message format"
const ErrConnectionFailure = "bad connection"

func PSQLTest(t *testing.T) (*PSQLTestStage, *PSQLTestStage, *PSQLTestStage) {
	proxyPort, err := getFreePort()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	stage := &PSQLTestStage{
		t:         t,
		psqlPort:  getContainerHostPort(postgresContainerName, postgresPort),
		proxyPort: proxyPort,
	}
	t.Cleanup(func() {
		if stage.db != nil {
			stage.db.Close()
		}

		proxy, _ := toxiclient.NewClient(fmt.Sprintf("localhost:%d", toxiProxyPort)).Proxy(proxyName)
		if proxy != nil {
			proxy.Delete()
		}
	})
	return stage, stage, stage
}

func (s *PSQLTestStage) and() *PSQLTestStage {
	return s
}

func (s *PSQLTestStage) a_connection_to_postgres() *PSQLTestStage {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", s.proxyPort, "postgres", "postgres", "postgres")

	db, err := sql.Open("postgres", psqlInfo)

	if err != nil {
		s.t.Error(err)
		s.t.Fail()
	}

	s.db = db
	return s
}

func (s *PSQLTestStage) a_psql_toxic(o options) *PSQLTestStage {
	client := toxiclient.NewClient(fmt.Sprintf("localhost:%d", toxiProxyPort))

	proxy, err := client.CreateProxy(proxyName,
		fmt.Sprintf("localhost:%d", s.proxyPort),
		fmt.Sprintf("localhost:%d", s.psqlPort))

	if err != nil {
		s.t.Error(err)
		s.t.Fail()
	}
	_, err = proxy.AddToxic(toxicName, "psql", "upstream", 100, map[string]interface{}{
		"FailAfter":    o.FailAfter,
		"SearchText":   o.SearchText,
		"RecoverAfter": o.RecoverAfter,
		"FailureType":  o.FailureType,
	})

	if err != nil {
		s.t.Error(err)
		s.t.Fail()
	}

	return s
}

func (s *PSQLTestStage) a_query_succeeds(q string) *PSQLTestStage {
	rows, err := s.db.Query(q)

	if err != nil {
		s.t.Error(err)
		s.t.Fail()
	}

	err = rows.Close()
	if err != nil {
		s.t.Error(err)
		s.t.Fail()
	}

	return s
}

func (s *PSQLTestStage) a_query_fails(q string, error string) *PSQLTestStage {
	rows, err := s.db.Query(q)
	s.t.Log(err)

	if err == nil || !strings.Contains(err.Error(), error) {
		rows.Close()
		s.t.Error("expected error but no error happened")
		s.t.Fail()
	}

	return s
}
