# regdb: 依赖注册和注入

## 使用示例

### 简单的内置对象的注册和注入
```go
var regDB RegDB

var out int = 100
regDB.RegisterObjectWithName("integer", out)

var in int
regDB.InjectObjectByName("integer", &in)
```

### 注册结构体对象，注入到接口
```go
var regDB RegDB

var out *bytes.Buffer
regDB.RegisterObjectWithName("reader", out)

var in io.Reader
regDB.InjectObjectByName("reader", &in)
```

### 对象内字段的注册
```go
var regDB RegDB

var out = struct {
    Integer int `out:"integer"`
}{
    Integer: 123,
}
regDB.RegisterObjectFields(out, "out")
```

### 注入到对象内字段
```go
var regDB RegDB

in := struct {
    Integer int `in:"integer"`
}{}

regDB.InjectObjectFields(&in, "in")
```

### 注入到对象内嵌套字段
```go
var regDB RegDB

type Nested struct {
    Integer int `in:"integer"`
}
in := struct {
    Nested *Nested `in:""`
}{
    &Nested{},
}
regDB.InjectObjectFields(&in, "in")
```