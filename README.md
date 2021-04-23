# protoc-go-sql
Modify from [proto-go-sql]https://github.com/travisjeffery/proto-go-sql

Generate sql.Scanner,driver.Valuer and gorm.io/gorm/migrate.GormDataTypeInterface implementations for your Protobufs.

## Example

We want the generated struct for this Person message to implement sql.Scanner and driver.Valuer so we can easily write and read it as protobuf from Message.

So we compile the person.proto file:

``` proto
syntax = "proto3";

option go_package = "/;proto";

message Person {
  message Hobbies {
    repeated string hobbies = 1;
  }

  string id = 1;
  Hobbies hobbies = 2;
  map<uint64, Hobbies> hobbies_map = 3;
}

```

And run:

``` sh
$ make sql
```

Generating this person.sql.go:

``` go
func (t *Person) Scan(val interface{}) error {
	return p.Unmarshal(val.([]byte), t)
}

func (t *Person) Value() (driver.Value, error) {
	if t == nil {
		t = &Person{}
	}
	return p.Marshal(t)
}

func (*Person) GormDBDataType(*gorm.DB, *schema.Field) string {
	return "blob"
}

func (t *Person_Hobbies) Scan(val interface{}) error {
	return p.Unmarshal(val.([]byte), t)
}

func (t *Person_Hobbies) Value() (driver.Value, error) {
	if t == nil {
		t = &Person_Hobbies{}
	}
	return p.Marshal(t)
}

func (*Person_Hobbies) GormDBDataType(*gorm.DB, *schema.Field) string {
	return "blob"
}
```

And we're done!

## License

MIT

---

- [guanmai.cn](https://www.guanmai.cn)
