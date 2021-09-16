# avro-schema-gen
Gen Avro schema from Go struct.

For example, from this struct
```go
type Pointer struct {
	in *int
}
```

avro-schema-gen will create a schema like this

```json
{
    "type":"record",
    "name":"Pointer",
    "fields":[
        {
            "name":"in",
            "default":null,
            "type":[
                "null",
                "long"
            ]
        }
    ]
}
```

## How to use
```go
import avroschema "github.com/badboyd/avro-schema-gen"

type Pointer struct {
	in *int
}

func main() {
    avroschema.Generate(Pointer{})
}
```
