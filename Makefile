.PHONY:sql
sql:
	go install && cd _example/proto && protoc --sql_out=. --go_out=. person.proto

.PHONY:clean
clean: 
	rm _example/proto/person.pb.go  && rm _example/proto/person.sql.go
