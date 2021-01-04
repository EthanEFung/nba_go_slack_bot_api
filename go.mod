module go_nba

go 1.13

replace go_nba/schedule => ./schedule

require (
	github.com/gorilla/mux v1.8.0
	go_nba/schedule v0.0.0-00010101000000-000000000000
)
