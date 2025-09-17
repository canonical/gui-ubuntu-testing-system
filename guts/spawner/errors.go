package main

type PostgresServiceNotUpError struct{}

func (e PostgresServiceNotUpError) Error() string {
	return "Unit postgresql.service is not active."
}
