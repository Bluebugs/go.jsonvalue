# 额外选项配置

[上一页](./05_caseless.md) | [总目录](./README.md) | [下一页](./07_conversion.md)

---

- [选项概述](./06_option.md#选项概述)
- [忽略 null 值](./06_option.md#忽略-null-值)
- [处理浮点数 NaN](./06_option.md#处理浮点数-nan)
- [处理浮点数 +/-Inf](./06_option.md#处理浮点数--inf)
- [旧版 options](./06_option.md#旧版-options)

---

## 选项概述

我们来回顾一下前文介绍的 `Marshal` 函数原型：

```go
func (v *V) Marshal(opts ...Option) (b []byte, err error)
func (v *V) MarshalString(opts ...Option) (s string, err error)
func (v *V) MustMarshal(opts ...Option) []byte
func (v *V) MustMarshalString(opts ...Option) string
```

可以看到每一个函数中都支持传入可选参数 `opts ...Option`。这些参数代表了将 JSON 序列化时的额外选项。

比如在 marshal 的时候忽略所有 object 中的 null 值，可以采用以下调用：

```go
v := jsonvalue.NewObject()
v.SetNull().At("null")
fmt.Println(v.MustMarshalString())
fmt.Println(v.MustMarshalString(jsonvalue.OptOmitNull(true)))
```

输出：

```json
{"null":null}
{}
```

各种选项说明如下：

---

## 忽略 null 值

参见上文

---

## 指定 key 顺序

在 jsonvalue 对 object 的实现是使用原生 `map` 类型实现的，因此在迭代每一个 kv 的时候，key 的顺序无法保证。

如果要固定 KV 的顺序，开发者可以在 marshal 的时候通过额外选项进行指定。

一般情况下，指定 KV 顺序是不必要的，而且还会增大序列化的开销。但也有一些特殊情况，有必要指定顺序：

- 需要对 JSON 的字节流进行哈希校验，因此需要保证同样的数据序列化后的字节流完全一致
- 调试期间便于快速找到指定的 K-V 对
- 对 JSON 的 object 进行了不规范的使用，对 key 的顺序有强要求

### 使用回调函数指定

```go
func OptKeySequenceWithLessFunc(f MarshalLessFunc) Option
```

`MarshalLessFunc` 是一个回调函数，定义如下：

```go
type MarshalLessFunc func(nilableParent *ParentInfo, key1, key2 string, v1, v2 *V) bool
```

参数定义如下：

- `nilableParent` - 表示当前层级 value 的父层级情况
- `key1`, `v1` - 表示需要排序的第一个 K-V 值
- `key2`, `v2` - 表示需要排序的第二个 K-V 值

回调函数中返回 bool 值表示 v1 是否应该在 v2 的前面。逻辑与 `sort` 包的 `Less` 函数逻辑相同。

在 jsonvalue 中也提供了一个最简单的回调函数，仅使用字母序进行排序：

```go
func OptDefaultStringSequence() Option
```

### 使用预定义的 []string 指定 key 顺序

使用 less 函数的格式较为复杂。开发者也可以简单提供一个 []string，这样 jsonvalue 则根据这个 string 顺序进行序列化。如果是在 `[]string` 中未指定的 key，则统一附在靠后的位置。

```go
func OptKeySequence(seq []string) Option
```

如果 `OptKeySequence` 和 `OptKeySequenceWithLessFunc` 同时指定，则优先采用 `OptKeySequenceWithLessFunc`。

--- 

## 处理浮点数 NaN

在 JSON 规范中，NaN（不是一个数字） 和 +/-Inf（正/负无穷）是非法的。但是在实际开发中，特别是数据科学和机器学习层面有时候有必要传递这个值。

在 jsonvalue 中针对 NaN 和 +/-Inf 值，默认逻辑是会进行报错并返回。但同时也提供了一些选项进行无错转换。首先我们看一下 NaN 的处理方案：

### NaN 转换成另一个浮点值

```go
func OptFloatNaNToFloat(f float64) Option
```

指定一个浮点数，当遇到 NaN 时，序列化时则替换成该浮点数。不能指定转换为 NaN 或 +/-Inf，否则会报错。

### 转换成 null

```go
func OptFloatNaNToNull() Option
```

当遇到 NaN 时，序列化时转换成 JSON null。注意，该选项不受 `OptOmitNull` 的影响，而是必然会将 NaN 转换成 null

### 转换成字符串

```go
func OptFloatNaNToString(s string) Option
func OptFloatNaNToStringNaN() Option
```

当遇到 NaN 时，序列化时替换成一个 string。`OptFloatNaNToStringNaN()` 等效于 `OptFloatNaNToString("NaN")`。

---

## 处理浮点数 +/-Inf

对于无穷，也有与 NaN 类似的处理方案：

### 转换成有效的浮点数

```go
func OptFloatInfToFloat(f float64) Option
```

指定一个浮点数，当遇到 +Inf 时，序列化时则替换成该浮点数；对于 -Inf 则替换成 `-f`。不能指定转换为 NaN 或 +/-Inf，否则会报错。

### 转换成 null

```go
func OptFloatInfToNull() Option
```

同样地，该选项不受 `OptOmitNull` 的影响，而是必然会将 +/-Inf 转换成 null

### 转换成字符串

```go
func OptFloatInfToString(positiveInf, negativeInf string) Option
func OptFloatInfToStringInf() Option
```

分别指定两个字符串，表示遇到 +/-Inf 时，分别替换成什么字符串。如果字符串为空，按照以下优先级替换：

- 对于 +Inf，如果配置字符串为空，则替换为默认的 `"+Inf"`
- 对于 -Inf，如果配置字符串为空，则首先查找 +Inf 配置，如果 +Inf 配置有的话，那么在其前缀上去掉 `+`，加上 `-` 符号
- 对于 -Inf，如果配置字符串为空，则替换为默认的 `"-Inf"`

`OptFloatInfToStringInf()` 函数等效于 `OptFloatInfToString("+Inf", "-Inf")`

## 旧版 options

在旧版本（v1.1.1 及以前）对 option 的使用方式是传入一个 `Opt` 类型的 struct。新版也兼容这种模式，但建议改为前文所述的函数模式进行配置。
