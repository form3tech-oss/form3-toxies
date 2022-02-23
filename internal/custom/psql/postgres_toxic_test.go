package psql

import (
	"testing"

	_ "github.com/lib/pq"
)

func TestNoFailure(t *testing.T) {
	given, _, then := PSQLTest(t)

	given.
		a_connection_to_postgres().and().
		a_psql_toxic(options{SearchText: ".*SELECT 1.*"})
	then.
		a_query_succeeds("SELECT 1").and().
		a_query_succeeds("SELECT 1").and().
		a_query_succeeds("SELECT 1")
}

func TestFailAfter(t *testing.T) {
	given, when, then := PSQLTest(t)

	given.
		a_connection_to_postgres().and().
		a_psql_toxic(options{FailOn: 3, SearchText: ".*SELECT 1.*"})
	when.
		a_query_succeeds("SELECT 1").and().
		a_query_succeeds("SELECT 1").and()
	then.
		a_query_fails("SELECT 1", ErrMessageFailure)
}

func TestFailConnectionAfter(t *testing.T) {
	given, when, then := PSQLTest(t)

	given.
		a_connection_to_postgres().and().
		a_psql_toxic(options{FailOn: 3, SearchText: ".*SELECT 1.*", FailureType: FailureTypeConnectionFailure})
	when.
		a_query_succeeds("SELECT 1").and().
		a_query_succeeds("SELECT 1").and()
	then.
		a_query_fails("SELECT 1", ErrConnectionFailure)
}

func TestRecoverAfter(t *testing.T) {
	given, when, then := PSQLTest(t)

	given.
		a_connection_to_postgres().and().
		a_psql_toxic(options{FailOn: 3, RecoverAfter: 3, SearchText: ".*SELECT 1.*"})
	when.
		a_query_succeeds("SELECT 1").and().
		a_query_succeeds("SELECT 1").and().
		a_query_fails("SELECT 1", ErrMessageFailure)
	then.
		a_query_succeeds("SELECT 1")
}

func TestFailAfterWithNonMatchingQueries(t *testing.T) {
	given, when, then := PSQLTest(t)

	given.
		a_connection_to_postgres().and().
		a_psql_toxic(options{FailOn: 3, RecoverAfter: 3, SearchText: ".*SELECT 1.*"})
	when.
		a_query_succeeds("SELECT 1").and().
		a_query_succeeds("SELECT 2").and().
		a_query_succeeds("SELECT 3").and().
		a_query_succeeds("SELECT 1")
	then.
		a_query_fails("SELECT 1", ErrMessageFailure)
}

func TestReconfigure(t *testing.T) {
	given, when, then := PSQLTest(t)

	given.
		a_connection_to_postgres().and().
		a_psql_toxic(options{FailOn: 1, SearchText: ".*SELECT 1.*"})
	when.
		a_query_fails("SELECT 1", ErrMessageFailure).and().
		the_toxic_is_reconfigured(options{FailOn: 2, SearchText: ".*SELECT 1.*"})
	then.
		a_query_succeeds("SELECT 1").and().
		a_query_fails("SELECT 1", ErrMessageFailure)
}
