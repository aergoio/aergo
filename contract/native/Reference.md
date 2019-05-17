Aergo Smart Contract Language Reference (draft)
=======================================

## Introduction

Aergo Smart Contract Lanaguage(이하 ASCL)은 Aergo를 위한 smart contract 전용 언어다. ASCL은 개발자로 하여금 보다 쉽고 편하게 smart contract를 개발할 수 있도록 설계되었으며, 기존의 언어보다 작지만 더욱 빠르고 강력한 성능을 목표로 한다.

ASCL은 다음과 같은 특성을 가진다.

* Strongly typed 언어다.
* (상속과 추상화를 제외한) object-oriented language다.
* C, Java, Go 등과 같은 일반적인 언어 구문을 지원한다.
* Database 접근을 위한 SQL 확장 구문을 지원한다.
* Blockchain 접근을 위한 blockchain 확장 구문을 지원한다.
* [WebAssembly](https://webassembly.org/)로 컴파일된다.

## Notation

이 문서에서 각 syntax는 [Extended Backus-Naur Form (EBNF)](https://en.wikipedia.org/wiki/Extended_Backus%E2%80%93Naur_form)를 사용하여 표기한다.

```
Usage               Notation
------------------  --------
definition              =
termination             ;
altenation              |
optional             [ ... ]
repetition           { ... }
grouping             ( ... )
terminal string      " ... "
special sequence     ? ... ?
exception               -
```

## Requirements

ASCL로 작성된 프로그램은 다음 사항을 유의해야 한다.

* [UTF-8](https://en.wikipedia.org/wiki/UTF-8) 인코딩으로 작성해야 한다.
* 대소문자를 구분한다.
* 변수는 forward declaration만 허용한다.
* 함수는 선언한 위치와 상관없이 호출할 수 있다.

> **TODO** hardware 혹은 OS 설명 필요

## Lexical elements

### Letters and digits

ASCL에서 사용할 수 있는 문자와 숫자는 다음과 같다.

<pre>
<a name="letter">letter</a>        = ? [a-zA-Z_] ? ;
<a name="uni_char">uni_char</a>      = ? 임의의 unicode 범위 문자 ? ;

<a name="decimal_digit">decimal_digit</a> = ? [0-9] ? ;
<a name="octal_digit">octal_digit</a>   = ? [0-7] ? ;
<a name="hex_digit">hex_digit</a>     = ? [0-9a-fA-F] ? ;
</pre>

### Comments

comments는 다음과 같은 두가지 형식을 지원한다.

* Single-line comments (// 으로 시작하여 그 줄의 끝까지)

```
// This is a single-line comment
```

* Multi-line comments (/* 으로 시작하여 */ 까지)

```
/*
 * This is a multi-line comment
 */
```

> **TODO** doxygen 처리가 완료되면 관련 내용 추가 필요

### Constants

상수는 프로그램내에서 절대 변하지 않는 값을 나타내며, boolean, integer, floating-point, string 타입 상수등이 있다.

#### Boolean literals

boolean constant는 <a href="#predefined_id">predefined identifier</a>로 true, false를 사용할 수 있다.

<pre>
<a name="bool_lit">bool_lit</a> = "true" | "false" ;
</pre>

#### Integer literals

integer constant는 8진수, 10진수, 16진수 형태로 표기할 수 있다.

<pre>
<a name="integer_lit">integer_lit</a> = <a href="#decimal">decimal</a> | <a href="#octal">octal</a> | <a href="#hexadecimal">hexadecimal</a> ;

<a name="decimal">decimal</a>     = ( <a href="#decimal_digit">decimal_digit</a> - "0" ) { <a href="#decimal_digit">decimal_digit</a>  } ;
<a name="octal">octal</a>       = "0" { <a href="#octal_digit">octal_digit</a> } ;
<a name="hexadecimal">hexadecimal</a> = "0" ( "x" | "X" ) <a href="#hex_digit">hex_digit</a> { <a href="#hex_digit">hex_digit</a> } ;
</pre>

```
10
2048
0700
0xff
0xABCD
```

#### Floating-point literals

floating-point constant는 소수점을 사용하거나 exponent를 사용하여 표기할 수 있다.

<pre>
<a name="float_lit">float_lit</a> = <a href="#natural">natural</a> "." [ <a href="#natural">natural</a> ] [ <a href="#exponent">exponent</a> ] |
            "." <a href="#natural">natural</a> [ <a href="#exponent">exponent</a> ] |
            <a href="#natural">natural</a> <a href="#exponent">exponent</a> ;

<a name="natural">natural</a>   = <a href="#decimal_digit">decimal_digit</a> { <a href="#decimal_digit">decimal_digit</a> } ;
<a name="exponent">exponent</a>  = ( "e" | "E" ) [ "+" | "-" ] <a href="#natural">natural</a> ;
</pre>

```
1.23
71.
.009
1E3
1e+19
3.14e-02
.4E+3
```

#### String literals

string constant는 쌍따옴표를 사용하여 표기할 수 있다.

<pre>
<a name="string_lit">string_lit</a> = "\"" { <a href="#uni_char">uni_char</a> - "\n" } "\"" ;
</pre>

```
""
"bicycle"
"traffic is jam"
```

### Identifiers

identifier는 변수나 함수, 타입의 이름으로 사용되며, 첫글자는 반드시 문자이거나 _(underscore)로 시작해야 한다. 그리고, identifier의 길이는 128bytes보다 작거나 같아야 한다.

<pre>
<a name="identifier">identifier</a> = <a href="#letter">letter</a> { <a href="#letter">letter</a> | <a href="#decimal_digit">decimal_digit</a> } ;
</pre>

```
abc
Var1
_case
player_cnt
```

### Keywords

keyword는 ASCL에서 미리 정의한 것으로 identifier로 사용할 수 없다.

#### Reserved words

다음은 기본 예약어와 SQL 예약어다.

```
Basic reserved words
-------------------------------------------------------------------
break       case        const       continue    default     else
enum        false       for         func        goto        if
in          new         null        payable     public      return
struct      switch      true        type
```

```
SQL reserved words
-------------------------------------------------------------------
alter       create      delete      drop        insert      replace
select      update
```

#### <a name="predefined_id">Predefined identifiers</a>

다음은 미리 정의한 identifier들이다.

```
Type identifiers
------------------------------------------------------------------
bool        byte
int         int8        int16       int32       int64       int128
float       double
string
map
account
cursor
```

```
Constant identifiers
------------------------------------------------------
true        false
```

```
Object identifiers
------------------------------------------------------
null        this
```

* null
null은 map, contract, interface등 object 타입 변수의 초기값으로 할당된다. ASCL에서 object 타입은 원칙적으로 비교 불가능하나, null과의 비교만 예외적으로 허용한다.

```
map(int, string) m = null;
if (m == null) {
    isNull = true;
}
else {
    isNull = false;
}
```

* this
this는 contract에 정의한 state variable을 참조할 때 사용한다. 보통 local variable이나 parameter와 이름이 중복될때 이를 구분하기 위해 사용하거나 혹은 코드의 의미를 더욱 명확하게 하기 위한 용도로 사용한다.

```
contract Testing {
    int testNo = 0;
    func runTest(int testNo) {     // assume that is passed testNo as 1
        ... testNo ...             // using parameter testNo (== 1)
        or
        ... this.testNo ...        // using state variable testNo (== 0)
    }
}
```

### Operators and punctuations

다음은 ASCL에서 사용할 수 있는 <a href="#operator">operator</a>다.

```
+           -           *           /           %           =
+=          -=          *=          /=          %=          &
|           ^           &=          |=          ^=          >>=
<<=         >>          <<          &&          ||          <
>           <=          >=          ==          !=          ++
--          ?           :           !
```

다음은 ASCL에서 사용할 수 있는 punctuation이다.

```
(           )           {           }           [           ]
;           ,           .
```

### Types

ASCL에서는 다음과 같은 타입을 지원한다.

<pre>
<a name="type">Type</a>           = <a href="#primitive_type">primitive_type</a> | <a href="#complex_type">ComplexType</a> ;
<a name="primitive_type">primitive_type</a> = "bool" | "byte" | "int" | "int8" | "int16" | "int32" | "int64" | "int128"
                 "float" | "double" | "string" | "account" | "cursor" ;
<a name="complex_type">ComplexType</a>    = <a href="#struct_decl">StructDecl</a> | <a href="#enum_decl">EnumDecl</a> | <a href="#map_decl">MapDecl</a> ;
</pre>

#### Boolean type

boolean type은 true와 false 값을 저장할 수 있는 타입이다.

```
bool isMale = true;
bool isReceived = false;
```

#### Numeric types

numeric type은 integer 타입과 floating-point 타입, 일부 타입의 alias를 사용할 수 있다.

```
Type      Description                Value range
--------  -------------------------  ------------------------------------------
byte      8-bit unsigned integer     0 ~ 255
int8      8-bit signed integer       -128 ~ 127
int16     16-bit signed integer      -32768 ~ 32767
int32     32-bit signed integer      -2147483648 ~ 2147483647
int64     64-bit signed integer      -9223372036854775808 ~ 9223372036854775807
int128    128-bit signed integer     -2^127 ~ (2^127 - 1)
```

```
Type      Description
--------  --------------------------------------
float     IEEE-754 32-bit floating-point numbers
double    IEEE-754 64-bit floating-point numbers
```

```
Type      Description
--------  ----------------
int       alias for int32
```

```
int8 i = 0;
float f = 3.141592;
```

#### String type

string type은 가변길이 문자열을 저장할 수 있는 타입이다. string 타입에 저장된 문자열은 immutable 속성을 가져 프로그램내에서 변경되지 않는다.

```
string v = "This is a string";
```

#### Struct type

struct type은 1개 이상의 field들의 집합으로 다음과 같이 선언한다. struct type은 다른 struct type을 field로 사용할 수 있다.

<pre>
<a name="struct_decl">StructDecl</a> = "type" <a href="#identifier">identifier</a> "struct" "{" { <a href="#field_decl">FieldDecl</a> ";" } "}" ;
<a name="field_decl">FieldDecl</a>  = <a href="#type">Type</a> <a href="#identifier">identifier</a> [ { <a href="#array_decl">ArrayDecl</a> } ] ;
</pre>

```
type User struct {
    int identifier;
    string name;
    string address;
}

type Customer struct {
    User user;          // nested struct
    int purchaseCount;
}
```

#### Enum type

enum type은 1개 이상의 constant element들의 집합으로 다음과 같이 선언한다. 각 element들은 선행 element보다 1이 증가된 값이 저장되고, 만약 첫번째 element에 별도의 값이 정의되지 않은 경우 첫번째 element엔 0이 저장되고, 차례로 1씩 증가된 값이 저장된다.

<pre>
<a name="enum_decl">EnumDecl</a>       = "enum" <a href="#identifier">identifier</a> "{" { <a href="#enum_decl">EnumeratorDecl</a> "," } "}" ;
<a name="enum_decl">EnumeratorDecl</a> = <a href="#identifier">identifier</a> [ "=" <a href="#decimal_digit">decimal_digit</a> ] ;
</pre>

```
enum Medal {
    GOLD,           // == 0
    SILVER,         // == 1
    BRONZE,         // == 2
}
enum City {
    Seoul,          // == 0
    Incheon,        // == 1
    Busan = 5,      // == 5
    Jeju,           // == 6
}
```

#### Object types

object 타입은 contract와 interface 타입이 있고, object 타입으로 정의된 variable은 초기화되지 않은 경우 null이 저장되고, <a href="#ctor_decl">constructor</a>를 사용하여 객체를 할당할 수 있다.

##### Map type

map type은 key, value를 쌍으로 갖는 hashmap이다. map에서 사용하는 key 타입은 반드시 <a href="#comparable_op">comparable 타입</a>이어야 하나, value 타입은 임의의 타입을 사용할 수 있다.

<pre>
<a name="map_decl">MapDecl</a> = "map" "(" <a href="#type">Type</a> "," <a href="#type">Type</a> ")" ;
</pre>

```
map(int, string)
map(int, User)
map(int, map(string, string))

map(map(double, int), string)   // raise error
```

map 타입은 <a href="#init_exp">initializer expression</a>이나 <a href="#alloc_exp">allocator expression</a>을 사용하여 초기화할 수 있다.


> **TODO** built-in 함수 구현 후 설명 필요

##### Contract type

contract 타입은 contract 객체를 저장한다. contract 타입은 <a href="#init_exp">initializer expression</a>이나 <a href="#alloc_exp">allocator expression</a>을 사용할 수 없고, <a href="#ctor_decl">constructor</a>를 호출하거나 다른 contract 변수 혹은 함수의 반환값을 사용하여 초기화해야 한다.

```
MyContract myCon = null;
MyContract myCon = new MyContract();
MyContract myCon = ContractBuilder.build();
```

##### Interface type

interface 타입은 interface를 구현한 contract 객체를 저장한다. contract 타입과 마찬가지로 <a href="#init_exp">initializer expression</a>이나 <a href="#alloc_exp">allocator expression</a>을 사용할 수 없고, <a href="#ctor_decl">constructor</a>를 호출하거나 함수의 반환값으로 contract를 받아와 초기화해야 한다. 단, interface 타입간의 assignment는 허용하지 않는다.

```
interface CommonItf { ... }
contract MyContract implements Common { ... }

CommonItf con = null;
CommonItf con = new MyContract();

CommonItf con2 = con;       // raise error
```

#### Account type

> **TODO** 구현 후 추가 필요

#### Cursor type

> **TODO** 구현 후 추가 필요

### Type conversions

ASCL은 strongly typed language로서 implicit type conversion을 지원하지 않으며 만약 operand의 타입이 서로 다를 경우엔 에러가 발생한다.

```
int16 i = 0;
int32 j = 1;
int32 k = i + j;    // raise error
float f = i;        // raise error
```

다만, 예외적으로 operand가 constant인 경우엔 implicit conversion을 허용하나, 다음과 같이 같은 유형의 타입이 아닌 경우엔 위와 마찬가지로 에러가 발생한다.

```
Constant  Comparable types
--------  ---------------------------------------
booleans  bool
integers  byte, int8, int16, int32, int64, int128
floats    float, double
strings   string
```

```
bool b = 1;         // raise error
int i = 1.23;       // raise error
float f = 1024;     // raise error
string s = 61;      // raise error
```

만약 explicit type conversion을 하고 싶은 경우엔 <a href="#cast_exp">cast expression</a>을 사용한다.

## Program structures

Smart contract는 다음과 같이 import 선언부와 interface 선언부, contract 선언부로 나눠진다.

<pre>
SmartContract = ( { <a href="#import_decl">ImportDecl</a> } | { <a href="#interface_decl">InterfaceDecl</a> } | { <a href="#contract_decl">ContractDecl</a> } ) ;
</pre>

### Imports

import는 외부 smart contract 파일을 참조하기 위해 사용하는 구문이다. import하기 위한 파일은 반드시 local storage에 있어야 하고, 절대경로를 사용하거나 현재 contract 파일이 존재하는 위치로부터의 상대경로를 사용할 수 있다.

<pre>
<a name="import_decl">ImportDecl</a> = "import" <a href="#string_lit">string_lit</a> ;
</pre>

```
import "ExternalContract"
import "../RootContract"
import "/home/CommonContract"
```

### Interfaces

interface는 함수의 specification을 작성하는 부분으로 다음과 같이 선언한다.

<pre>
<a name="interface_decl">InterfaceDecl</a> = "interface" <a href="#identifier">identifier</a> "{" { <a href="#func_spec">FunctionSpec</a> } "}" ;
</pre>

```
interface Trade {
    public func propose(int price, int64 amount) bool;
    public func deal;
}
```

interface는 다음과 같은 제약조건을 가진다.

* <a href="#ctor_decl">constructor</a>는 정의할 수 없다.
* <a href="#func_body">function body</a>는 정의할 수 없다.
* <a href="#variable_decl">variable</a>이나 <a href="#struct_decl">struct</a>, <a href="#enum_decl">enum</a>등은 정의할 수 없다.


### Contracts

contract는 실제 로직을 작성하는 부분으로 다음과 같이 선언한다.

<pre>
<a name="contract_decl">ContractDecl</a> = "contract" <a href="#identifier">identifier</a> [ <a href="#impl_opt">Implements</a> ] "{" <a href="#contract_body">ContractBody</a> "}" ;
<a name="contract_body">ContractBody</a> = { <a href="#variable_decl">VariableDecl</a> | <a href="#struct_decl">StructDecl</a> | <a href="#ctor_decl">ConstructorDecl</a> | <a href="#func_decl">FunctionDecl</a> } ;
</pre>

```
contract Buyer { ... }
contract Seller { ... }
```

#### Contract instantiations

contract를 사용하기 위해선 반드시 constructor를 호출하여 먼저 객체화 해야 한다.

<pre>
<a name="instantiation">Instantiation</a> = "new" "identifier" "(" [ <a href="#arg_list">ArgumentList</a> ] ")" ;
</pre>

```
Buyer Tom = new Buyer();
Seller Jane = new Seller();

Tom.pay(Jane, 1000);
Jane.transfer(Tom, "items");
```

#### Interface implementations

contract는 다음과 같이 interface를 구현할 수 있다.

<pre>
<a name="impl_opt">Implements</a>   = "implements" <a href="#identifier">identifier</a> ;
</pre>

```
interface Trade { ... }

contract Buyer implements Trade { ... }
contract Seller implements Trade { ... }
```

interface를 implmenetation한 경우 다음의 규칙을 지켜야 하고, 그렇지 않은 경우엔 compile error가 발생한다.

* interface에 선언된 모든 함수를 반드시 정의해야 한다.
* interface에 선언된 함수와 specification이 일치해야 한다.

```
interface Trade {
    public func propose(int price, int64 amount) bool;
    public func deal;
}

contract Buyer implements Trade {
    public func propose(int price, int64 amount) bool { ... }
}
```

위 예제는 deal function을 구현하지 않았으므로 compile error가 발생한다.

> 현재는 하나의 contract에서 하나의 interface만 구현할 수 있으나, 이는 추후에 확장될 수 있다.

### Variables

variable은 값을 저장하기 위한 공간으로 다음과 같이 선언한다.

<pre>
<a name="variable_decl">VariableDecl</a>   = [ <a href="#qualifier">qualifier</a> ] <a href="#type">Type</a> <a href="#var_list">VariableList</a> [ "=" ( <a href="#init_list">InitalizerList</a> | <a href="#alloc_list">AllocatorList</a> ) ] ";" ;

<a name="var_list">VariableList</a>   = <a href="#variable">Variable</a> { "," <a href="#variable">Variable</a> } ;
<a name="variable">Variable</a>       = <a href="#identifier">identfier</a> [ <a href="#array_decl">ArrayDecl</a> ] ;

<a name="init_list">InitalizerList</a> = <a href="#init_exp">Initializer</a> { "," <a href="#init_exp">Initializer</a> } ;
<a name="alloc_list">AllocatorList</a>  = <a href="#alloc_exp">Allocator</a> { "," <a href="#alloc_exp">Allocator</a> } ;
</pre>

#### Variable declarations

variable은 타입과 이름을 차례로 나열하여 선언할 수 있다.

```
bool isMale;
int userId;
double profitRate;
string address;
```

또한, 다음과 같이 하나의 타입에 대응되는 2개 이상의 variable도 선언할 수도 있다.

```
int userId, deptId;
string name, deptName, grade;
```

단, 서로 다른 타입의 variable을 같은 라인에 선언하는 것은 불가능하다.

```
int identifier, string name;    // raise error
```

#### Variable scope

variable은 <a href="#block_stmt">block</a>을 기준으로 scope가 결정된다. 먼저 variable을 선언하는 경우 같은 block 내부에선 모든 variable의 이름이 unique해야 하고, 만약 중복될 경우엔 에러가 발생한다. 단, block이 중첩된 경우엔 서로 다른 block으로 취급하여 같은 이름을 사용하는 것을 허용한다.

```
int counter;
int counter;        // raise error
{
    int counter;    // ok
}
```

다음으로 variable을 참조하는 경우엔 현재 block을 포함한 모든 상위 block의 변수를 참조할 수 있다. 즉, k가 속한 block에서는 상위 block에 선언된 i, j를 모두 참조할 수 있다.

```
int i;
{
    int j;
    {
        int k = i + j;
    }
}
```

마지막으로 variable은 선언 순서와 위치에 따라 참조 가능 여부가 결정되는데, variable을 참조하기 위해선 반드시 선행되어 선언되어 있어야 한다. 다음 예에서 j는 선언되기 전에 참조하였으므로 에러가 발생한다.

```
int i;
i = 0;          // ok
j = 1;          // out of scope
int j;

```

#### Type qualifiers

type qualifier는 state variable의 속성을 정의하는 것으로 다음과 같이 부여할 수 있다.

<pre>
<a name="qualifier">qualifier</a> = [ "public" ] [ "const" ] ;
</pre>

```
Qualifier   Descriptions
---------   -----------------
public      exported variable
const       constant variable
```

프로그램에서 정의한 모든 state variables, constants, funtions(constructor는 제외) 등은 기본적으로 외부 smart contract에서 참조할 수 없는 private 속성을 가진다. 하지만, public qualifier가 주어진 경우엔 exported symbol로 간주되어 임의의 smart contract에서 참조할 수 있다.

```
public int x;       // 외부 참조 가능
int y;              // 외부 참조 불가능
```

> **TODO** storage 옵션이 추가되면 설명 필요

const 속성은 해당 변수에 constant 속성을 부여하는 것으로 최초 선언 이후엔 immutable 속성을 가진다. 만약 constant variable 선언시에 값을 지정하기 않은 경우엔 에러가 발생하고, 이후에 새로운 값을 할당하려고 할 경우에도 에러가 발생한다.

```
const int i = 0;
const int j;        // raise error
i = 1;              // raise error
```

#### Array declarations

모든 variable은 다음 형식을 사용하여 [Array](https://en.wikipedia.org/wiki/Array_data_type)로 선언할 수 있다.

<pre>
<a name="array_decl">ArrayDecl</a> = "[" <a href="#expression">Expression</a> "]" ;
</pre>

array를 선언하기 위해선 크기를 지정해야 하는데, 이 값은 반드시 0보다 크거나 같은 integer constant여야 한다. 단, 예외적으로 <a href="#init_exp">initializer expression</a>이나 <a href="#alloc_exp">allocator expression</a>이 정의된 경우엔 array 크기를 생략할 수 있다.

또한, array는 N-dimension으로 선언할 수 있다.

```
int i[10];
int j[3][4];
string s[1 + 1];
map(int64, double) m[2];

const int MAX_SIZE = 16;
int cars[MAX_SIZE];
```

array의 element에 접근하기 위해선 <a href="#index_exp">index expression</a>을 사용해야 하고, 첫번째 element부터 마지막 element까지 차례로 0 ~ (size of array - 1) 범위의 index를 사용한다.

### Expressions

expression은 operator와 operands의 결합으로 이뤄진다.

<pre>
<a name="exp">Expression</a> = <a href="#unary_exp">UnaryExp</a> | <a href="#bin_exp">BinaryExp</a> | <a href="#tern_exp">TernaryExp</a> | <a href="#tuple_exp">TupleExp</a> | <a href="#sql_exp">SqlExp</a> ;

<a name="unary_exp">UnaryExp</a>   = <a href="#primary_exp">PrimaryExp</a> | <a href="#unary_op">unary_op</a> <a href="#unary_exp">UnaryExp</a> ;

<a name="bin_exp">BinaryExp</a>  = <a href="#exp">Expression</a> <a href="#bin_op">binary_op</a> <a href="#exp">Expression</a> ;
<a name="bin_op">binary_op</a>  = <a href="#arith_op">arith_op</a> | <a href="#logical_op">logical_op</a> | <a href="#comparable_op">comp_op</a> | <a href="#bit_op">bitwise_op</a> ;

<a name="tern_exp">TernaryExp</a> = <a href="#exp">Expression</a> "?" <a href="#exp">Expression</a> ":" <a href="#exp">Expression</a> ;

<a name="tuple_exp">TupleExp</a>   = <a href="#exp">Expression</a> { "," <a href="#exp">Expression</a> } ;
</pre>

#### Primary expressions

primary expression은 unary, binary, ternary expression의 operand로 사용된다.

<pre>
<a name="primary_exp">PrimaryExp</a> = <a href="#id_exp">IdExp</a> | <a href="#val_exp">ValueExp</a> | <a href="#cast_exp">CastExp</a> | <a href="#index_exp">IndexExp</a> | <a href="#call_exp">CallExp</a> | <a href="#access_exp">AccessExp</a> | <a href="#init_exp">InitExp</a> | <a href="#alloc_exp">AllocExp</a> |
             <a href="#primary_exp">PrimaryExp</a> "++" | <a href="#primary_exp">PrimaryExp</a> "--" | "(" <a href="#exp">Expression</a> ")" ;
</pre>

##### Identifier expressions

identifier expression은 이름을 참조하기 위한 것으로 variable, struct, enum, function, contract, interface 등의 user-defined identifier와 predefined identifier를 사용할 수 있다. 만약 존재하지 않는 이름을 사용할 경우엔 에러가 발생한다.

<pre>
<a name="id_exp">IdExp</a> = <a href="#identifier">identifier</a> ;
</pre>

##### Value expressions

value expression은 boolean, integer, floating point, string constant다.

<pre>
<a name="val_exp">ValueExp</a> = <a href="#bool_lit">bool_lit</a> | <a href="#integer_lit">integer_lit</a> | <a href="#float_lit">float_lit</a> | <a href="#string_lit">string_lit</a> ;
</pre>

```
true
false
1
2.984
"rainbow"
```

##### Cast expressions

cast expression은 <a href="#primitive_type">primitive type</a>대한 explicit type conversion을 정의한다.

<pre>
<a name="cast_exp">CastExp</a> = "(" <a href="#primitive_type">primitive_type</a> ")" <a href="#primary_exp">PrimaryExp</a>
</pre>

```
int i = 0;
bool b = (bool)i;
double d = (double)i;
string s = (string)i;
```

<a href="#complex_type">complex type</a>에 대한 type conversion은 지원하지 않는다.

```
map(int, string) m1;
map(int, string) m2 = (map(int, string))m1;     // raise error
```

##### Index expressions

index expression은 array element에 접근하기 위한 것으로 index는 반드시 0보다 크거나 같은 양의 정수여야 한다.

<pre>
<a name="index_exp">IndexExp</a> = <a href="#primary_exp">PrimaryExp</a> "[" <a href="#exp">Expression</a> "]" ;
</pre>

```
ids[0]
counts[1 + 2 * 3]
names[-1]           // raise error
names[2.1]          // raise error
```

##### Call expressions

call expression은 <a href="#ctor_decl">constructor</a>나 <a href="#func_decl">function</a>을 호출하기 위해 사용한다.

<pre>
<a name="call_exp">CallExp</a>      = <a href="#primary_exp">PrimaryExp</a> "(" [ <a href="#arg_list">ArgumentList</a> ] ")" ;
<a name="arg_list">ArgumentList</a> = <a href="#exp">Expression</a> { "," <a href="#exp">Expression</a> } ;
</pre>

argument list는 0개 이상의 expression으로 구성되며, 각 argument들은 function의 specification과 정확히 같은 타입을 가져야 하고, 그렇지 않을 경우엔 에러가 발생한다.

```
// func f (int16 p1, string p2) ...

int16 i = 0;
int32 j = 0;
string s = "lemon";

f(i, s);            // ok
f(j, s);            // raise error
```

일반적으로 argument가 parameter로 전달될 때는 값이 복사되어 전달되지만, 다음의 타입들에 대해선 reference가 전달된다.

* array
* struct
* map
* contract

##### Access expressions

access expression은 struct field에 접근하거나 contract state variable을 참조하거나 혹은 함수를 호출하기 위해 사용한다.

<pre>
<a name="access_exp">AccessExp</a> = <a href="#primary_exp">PrimaryExp</a> "." <a href="#identifier">identifier</a> ;
</pre>

```
type coord struct {
    int x;
    int y;
    int z;
}
coord calc;
calc.x = 1;
calc.y = 1;
calc.z = calc.x + calc.y;
```

##### Initializer expression

initializer expression은 complex variable 값을 정의할 때 사용하며, "new {" ... "}" 형태로 표현한다.

<pre>
<a name="init_exp">InitExp</a>     = "new" <a href="#initializer">Initializer</a> ;
<a name="initializer">Initializer</a> = "{" <a href="#elem_list">ElementList</a> "}" ;
<a name="elem_list">ElementList</a> = ( <a href="#expression">Expression</a> | <a href="#initializer">Initializer</a> ) { "," ( <a href="#expression">Expression</a> | <a href="#initializer">Initializer</a> ) } ;
</pre>

다음은 array initializer다. multi-dimension인 경우엔 중첩해서 선언한다.

```
int levels[2] = new { 98, 99 };
string classes[3] = new { "magician", "barbarian", "archer" };

const int DIVISION = 2;
const int WEEK = 7;
double salesPerWeek[DIVISION][WEEK] = new {
    // first division
    { 1.3, 2.0, 2.1, 0.3, 1.8, 6.4, 5.7 },
    // second division
    { 1.7, 3.8, 2.1, 1.1, 7.3, 5.0, 2.5 },
};
```

다음은 struct, map variable에 대한 initializer다.

```
type Singer struct {
    string name;
    int debutYear;
    int albumCnt;
}
Singer michael = new { "Michael Jackson", 1964, 11 };
Single kpops[2] = new {
    { "Yong-pil Cho", 1979, 19 },
    { "Mi-ja Lee", 1959, 500 },
};

map(int, string) keystore = new {
    { 20180850, "19ffbaae54a4c1c4cd6ceef01eff0595e3c778ce" },
    { 20181025, "3b6af44ef92fb973626925ddfd79a77dcd70456e" },
    { 20181357, "043ade3730e2172d917575132dff58f271ad59f4" },
};
```

##### Allocator expression

allocator expression은 array나 struct, map의 메모리 공간을 할당할 때 사용하며, "new type[...]" 형태로 표현한다.

<pre>
<a name="alloc_exp">AllocExp</a> = "new" <a href="#type">Type</a> [ { <a href="#array_decl">ArrayDecl</a> } ] ;
</pre>

다음은 array allocator다. multi-dimension인 경우엔 중첩해서 선언한다.

```
int levels[2] = new int[2];
int64 classes[] = new int64[3];
double areas[][] = new double[4][5];
```

다음은 struct, map variable에 대한 allocator다.

```
type Game struct {
    int8 category;
    string name;
}
Game lol = new Game;
Game shooting[2] = new Game[2];

map(int, string) actors = new map(int, string);
map(int, string) cities[5] = new map(int, string)[5];
```

#### <a name="operator">Operators</a>

operator는 크게 unary, binary, ternary 세가지로 이뤄지고, binary operator는 그 특성에 따라 다시 몇가지로 분류할 수 있다.

##### Unary operators

unary operator는 단일 expression에 적용되어 같은 타입의 값을 반환한다.

<pre>
<a name="unary_op">unary_op</a> = "++" | "--" | "+" | "-" | "!" ;
</pre>

```
Operator  Description                 Applicable types
--------  --------------------------  ----------------
++        increase and assign         integer, float
--        decrease and assign         integer, float
+         positive                    integer, float
-         negative                    integer, float
!         logical NOT                 boolean
~         bitwise NOT                 integer
```

* \+ operator는 양수임을 표현할 뿐 값에는 아무런 영향을 끼치지 않는다.
* \- operator는 값의 부호를 바꾼다. 즉, 양수는 음수로 음수는 양수로 바꾼다.
* ++, -- operator는 constant에는 사용할 수 없다.

##### Arithmetic operators

arithmetic operator는 binary operator로 정수나 부동소수 혹은 string을 이용하여 연산을 수행하고 동일한 타입의 값을 반환한다.

<pre>
<a name="arith_op">arith_op</a> = "+" | "-" | "*" | "/" | "%" ;
</pre>

```
Operator  Description                 Applicable types
--------  --------------------------  -------------------------
+         add                         integers, floats, strings
-         subtract                    integers, floats
*         multiply                    integers, floats
/         divide                      integers, floats
%         remainder                   integers
```

* \+ operator가 string 타입에 적용될 경우 concatenated string을 만든다.

```
string str = "1" + "2";     // str is "12"
```

* /, % operator의 right expression의 값이 0인 경우엔 에러가 발생한다.

##### Logical operators

logical operator는 binary operator로 좌측 operand부터 차례로 true, false를 판단하여 bool 타입의 값을 반환한다.

<pre>
<a name="logical_op">logical_op</a> = "&&" | "||" ;
</pre>

```
Operator  Description                 Applicable types
--------  --------------------------  ----------------
&&        logical AND                 booleans
||        logical OR                  booleans
```

##### Comparison operators

comparison operator는 binary operator로 operands의 값을 비교하여 bool 타입의 값을 반환한다.

<pre>
<a name="comparable_op">comp_op</a> = "==" | "!=" | "<" | ">" | "<=" | ">=" ;
</pre>

```
Operator  Description                 Applicable types
--------  --------------------------  --------------------------------
==        equals                      bools, integers, floats, strings
!=        not equals                  bools, integers, floats, strings
<         less than                   bools, integers, floats, strings
>         greater than                bools, integers, floats, strings
<=        less than or equals         bools, integers, floats, strings
>=        greater than or equals      bools, integers, floats, strings
```

* bool, integer, float, string 타입은 comparable이다.
* comparable 타입의 array는 comparable이다.
* comparable 타입으로 이뤄진 struct는 comparable이다.
* map, contract 타입은 comparable하지 않다.

예외적으로 map 타입과 contract 타입 variable은 predefined identifier인 null과 비교할 수 있다.

##### Bitwise operators

bitwise operator는 binary operator로 각 operands의 bit 값을 비교 혹은 계산하여 integer 타입의 값을 반환한다.

<pre>
<a name="bit_op">bitwise_op</a> = "&" | "|" | "^" | ">>" | "<<" ;
</pre>

```
Operator  Description                 Applicable types
--------  --------------------------  --------------------------
&         bitwise AND                 integers
|         bitwise OR                  integers
^         bitwise XOR                 integers
>>        right shift                 integer >> unsigned integer
<<        left shift                  integer << unsigned integer
```

* 모든 operand는 반드시 integer 타입이어야 한다.
* shift operator의 right operand는 반드시 unsigned 값이어야 한다.

##### Ternary operators

ternary operator는 3개의 operand를 사용하여 첫번째 conditional operator의 값에 따라 두번째 expr1 값이나 세번째 expr2 값을 반환한다.

```
variable = condition ? expr1 : expr2;
```

따라서 위와 같은 statement는 단순히 다음과 같은 형태로 바뀔 수 있다.

```
if (condition) {
    variable = expr1;
} else {
    variable = expr2;
}
```

또한, ternary operator는 각 위치에 중첩되어 사용될 수 있다.

```
(text == "first" ? 1 : (text == "second" ? 2 : 3)
```

##### Operator precedence

각 operator 사이엔 우선순위가 있으며, 아래에 나열한 순서대로 우선순위가 결정된다. 우선순위가 높을수록 강한 결합력을 가지므로 우선순위가 낮은 operator보다 먼저 적용된다.

```
Precedence  Operator
----------  -----------
    1       * / % >> <<
    2       + -
    3       <= >= < >
    4       == !=
    5       & ^ | !
    6       &&
    7       ||
```

```
a + b * c           // equals a + (b * c)
a + b > c           // equals (a + b) > c
a ^ b && c | d      // equals (a ^ b) && (c | d)
a && b || c         // equals (a && b) || c
```

위 precedence table에서 ++와 -- operator가 빠져있는데, 두 operator는 선언된 위치에 따라 동작이 달라지기 때문이다. 만약 prefix로 선언된 경우엔 가장 높은 precedence를 가지나, postfix로 선언된 경우엔 가장 낮은 precedence를 가진다.
```
int i = 1;
int j = 10 + ++i;   /* first, execute i = i + 1
                       second, execute j = 10 + i
                       so, j is 12 and i is 2 */
int j = 10 + i++;   /* first, execute j = 10 + i
                       second, execute i = i + 1
                       so, j is 11 and i is 2 */
```

### Statements

statement는 variable과 constant, expression등의 조합으로 이뤄지며, 일종의 atomic 객체로 취급하여 개별 expression 수행중에 에러가 발생하는 경우엔 즉시 수행이 중지된다.

<pre>
<a name="stmt">Statement</a> = <a href="#exp_stmt">ExpressionStmt</a> | <a href="#assign_stmt">AssignStmt</a> | <a href="#label_stmt">LabelStmt</a> | <a href="#if_stmt">IfStmt</a> | <a href="#loop_stmt">LoopStmt</a> |
            <a href="#switch_stmt">SwitchStmt</a> | <a href="#exp_stmt">JumpStmt</a> | <a href="#exp_stmt">DdlStmt</a> | <a href="#exp_stmt">BlockStmt</a> ;
</pre>

#### Expression statements

expression statement는 expression으로만 이뤄지며, ; 만 있는 경우엔 empty statement로 아무런 동작을 하지 않는다.

<pre>
<a name="exp_stmt">ExpressionStmt</a> = ";" | <a href="#exp">Expression</a> ";" ;
</pre>

```
;
gift = "KinderJoy";
```

#### Assignement statements

assignment statement는 right operand의 값을 left operand에 저장한다.

<pre>
<a name="assign_stmt">AssignStmt</a> = <a href="#exp">Expression</a> <a href="#assign_op">assign_op</a> <a href="#exp">Expression</a> ";" ;
<a name="assign_op">assign_op</a>  = "=" | "+=" | "-=" | "*=" | "/=" | "&=" | "|=" | "^=" | ">>=" | "<<=" ;
</pre>

```
Operator  Description                 Applicable types
--------  --------------------------  -------------------------
=         assign                      any types
+=        add and assign              integers, floats, strings
-=        subtract and assign         integers, floats
*=        multiply and assign         integers, floats
/=        divide and assign           integers, floats
&=        bitwise AND and assign      integers
|=        bitwise OR and assign       integers
^=        bitwise XOR and assign      integers
```

* = operator는 각 operand에 <a href="#tuple_exp">TupleExp</a>를 사용할 수 있다.

```
i = 0;
x, y = 1, 10.78;
id, name = getUserInfo();
```

* composite operator들은 단일 expression만 허용한다.

```
x, y += 1, 1;       // raise error
```

* composite operator들은 <a href="#arith_op">arithmetic operator</a>와 <a href="#bit_op">bitwise operator</a> 규칙을 따른다.

#### Label statements

label statement는 <a href="#jump_stmt">goto</a>의 대상이다.

<pre>
<a name="label_stmt">LabelStmt</a> = <a href="#identifier">identifier</a> ":" <a href="#stmt">Statement</a> ;
</pre>

```
Increase: count++;
```

#### If statements

if statement는 conditional statement라고도 불리며, boolean condition에 따라 분기하여 수행된다. 만약 condition이 true일 경우엔 if branch가 수행되고, false일 경우엔 else branch가 수행된다.

<pre>
<a name="if_stmt">IfStmt</a> = "if" "(" <a href="#exp">Expression</a> ")" <a href="#block_stmt">BlockStmt</a> [ "else" ( <a href="#if_stmt">IfStmt</a> | <a href="#block_stmt">BlockStmt</a> ) ] ;
</pre>

```
if (scroe >= 90) {
    rank = "first";
}
else if (score >= 80) {
    rank = "second";
}
else {
    // nothing to do
}
```

#### Loop statements

loop statement는 조건에 따라 block을 반복 수행하기 위해 사용하거나 array 혹은 map iterator로 사용된다.

<pre>
<a name="loop_stmt">LoopStmt</a>      = <a href="#for_loop">ForLoopStmt</a> | <a href="#array_loop">ArrayLoopStmt</a> ;

<a name="for_loop">ForLoopStmt</a>   = "for" <a href="#block_stmt">BlockStmt</a> |
                "for" "(" <a href="#exp">Expression</a> ")" <a href="#block_stmt">BlockStmt</a> |
                "for" "(" <a href="#init_exp">InitExp</a> [ <a href="#exp_stmt">ExpressionStmt</a> ] [ <a href="#exp">Expression</a> ] ")" <a href="#block_stmt">BlockStmt</a> ;
<a name="init_exp">InitExp</a>       = <a href="#exp_stmt">ExpressionStmt</a> | <a href="#variable_decl">VariableDecl</a> ;

<a name="array_loop">ArrayLoopStmt</a> = "for" "(" <a href="#init_exp">InitExp</a> "in" <a href="#primary_exp">PrimaryExp</a> ")" <a href="#block_stmt">BlockStmt</a> ;
</pre>

##### Without condition

아무런 expression 없이 정의한 경우엔 항상 inifinite loop이다.

```
for {
    i = 1;
}
```

##### With single condition

단일 expression을 사용한 경우엔 해당 condition이 true인 동안 block을 반복 수행한다. 이 때 condition은 block 수행이 끝난 후에 다시 evaluation된다.

```
for (i < 10) {
    i++;
}
```

##### Traditional for-loops

traditional for-loop은 initialization, condition, afterthought의 세가지 부분으로 구성이 되고, 각 부분은 생략할 수 있다.

* Initialization
variable에 초기값을 할당하거나 variable을 직접 선언하여 초기값을 할당한다.
* Condition
boolean값을 반환하는 expression을 평가하여 block 수행여부를 결정한다.
* Afterthought
한번의 block 수행이 끝난 이후마다 반복 실행된다.

만약 condition이 생략되었고, block 내부에 break문이 없는 경우 혹은 세가지 부분이 모두 생략된 경우엔 infinite loop이 된다.

```
for ( ; ; ) { }
for (i = 0; ; ) { }
```

다음은 일반적인 for-loop이다.

```
for (i = 0; i < 10; ) {
    i += 2;
}
for (i = 0; i < 10; i++) { ... }
for (int j = 10; j >= 0; j--) { ... }
```

##### Iterator-based for-loops

iterator-based for-loop은 array나 map에 대한 iteration을 지원한다. 따라서 in 뒤에 오는 value expression은 반드시 array, map variable이거나 이 값을 반환하는 함수이어야 한다.

array의 경우엔 0번째 index element부터 시작하여 마지막 index element까지 반복한다.

```
int ids[5] = { 1, 2, 3, 4, 5 };

for (int id in ids) {
    // iterate over 1, 2, 3, 4, 5 in turn
}
```

map의 경우엔 모든 key, value 쌍에 대해 반복한다. 단, 이 경우엔 element의 순서가 보장되지 않는다.

```
map(int, string) users = {
    { 1, "James" }, { 2, "Sam" }, { 3, "Tom" }
};

int id;
string name;
for (id, name in users) {
    // iterate over randomly
}
```

#### Switch statements

switch statement는 if statement와 유사하게 condition을 평가하여 각 branch로 분기한다. condition은 switch clause에 정의하거나 case clause에 정의해야 한다.

<pre>
<a name="switch_stmt">SwitchStmt</a>     = "switch" [ "(" <a href="#exp">Expression</a> ")" ] "{" { <a href="#case_list">CaseClauseList</a> } [ <a href="#dflt_clause">DefaultClause</a> ] "}" ;
<a name="case_list">CaseClauseList</a> = "case" <a href="#exp">Expression</a> ":" { <a href="#stmt">Statement</a> } ;
<a name="dflt_clause">DefaultClause</a>  = "default" ":" { <a href="#stmt">Statement</a> } ;
</pre>

switch condition에 value expression이 존재하는 경우 제일 위 case clause에 정의된 value부터 비교하여 일치하면 해당 statement를 수행한다. 만약 일치하는 case clause가 없고, default clause가 있는 경우엔 default statement가 수행된다. default clause도 없는 경우엔 아무런 동작을 하지 않는다.

```
switch (marketCap) {
case 1:
    currency = "Bitcoin";
    break;
case 2:
    currency = "Ethereum";
    break;
default:
    currency = "Unknown";
}
```

하나의 case clause는 반드시 <a href="#jump_stmt">break statement</a>로 끝나야 하고, 그렇지 않을 경우 다음 case statement를 수행하게 된다. 단, default clause는 break를 생략할 수 있다. 다음 예제에서 type이 1인 경우 name은 "int"가 아닌 "string"이 된다.

```
switch (type) {
case 1:
    name = "int";
case 2:
    name = "string";
    break;
}
```

switch condition이 비어있고, case condition에 boolean expression이 있는 경우엔 위와 마찬가지로 제일 위 case clause의 boolean condition부터 평가하여 true를 반환하는 경우 해당 statement를 수행하게 된다.

```
switch {
case color == "Red":
    code = 0xFF0000;
    break;
case color == "Blue":
    code = 0x0000FF;
    break;
default:
    code = 0x000000;
}
```

#### Jump statements

jump statement는 코드의 순차적 수행을 벗어나 loop의 처음으로 돌아가거나(continue) loop을 벗어나거나(break) 현재의 function block을 벗어나거나(return) 특정 위치로 이동하는(goto) 경우에 사용한다.

<pre>
<a name="jump_stmt">JumpStmt</a> = "continue" ";" |
           "break" ";" |
           "return" [ <a href="#exp_list">ExpressionList</a> ] ";" |
           "goto" <a href="#identifier">identifier</a> ";" ;
</pre>

continue statement는 현재의 loop 처음으로 이동하여 다음 iteration을 수행한다.

```
int j = 0;
for (int i = 0; i < 10; i++) {  // execute 10 times
    j++;
    continue;
    j--;                        // never be executed
}
// j is 10
```

break statement는 현재 loop을 벗어난다.

```
int j = 0;
for (int i = 0; i < 10; i++) {  // execute 1 times
    j++;
    break;
    j--;                        // never be executed
}
// j is 1
```

return statement는 현재의 <a href="#func_decl">function block</a>을 벗어난다.

```
func f() {
    return;
    int i = 0;          // never be executed
}
```

goto statement는 label 위치로 이동한다.

```
goto final;
value = 0;              // never be executed
final:
    value = 1;
```

#### Block statements

block statement는 { }로 감싸진 variables, structs, statements 등의 집합이다.
<pre>
<a name="block_stmt">BlockStmt</a> = "{" { <a href="#variable_decl">VariableDecl</a> | <a href="#struct_decl">StructDecl</a> | <a href="#stmt">Statement</a> } "}" ;
</pre>

### Constructors

constructor는 하나의 smart contract를 객체화하기 위해 사용하는 함수로 항상 public 속성을 가진다. 특히 external smart contract를 사용하기 위해서는 반드시 해당 contract를 객체화해야 한다.

<pre>
<a name="ctor_decl">ConstructorDecl</a> = <a href="#identifier">identifier</a> "(" [ <a href="#param_list">ParameterList</a> ] ")" <a href="#block_stmt">BlockStmt</a> ;
<a name="param_list">ParameterList</a>   = <a href="#param_decl">ParameterDecl</a> { "," <a href="#param_decl">ParameterDecl</a> } ;
<a name="param_decl">ParameterDecl</a>   = [ "const" ] <a href="#type">Type</a> <a href="#identifier">identifier</a> ;
</pre>

#### Constructor declarations

constructor는 생략가능하나 그렇지 않은 경우 반드시 1개만 있어야 하고, constructor의 이름은 반드시 contract 이름과 일치해야 한다.

```
contract Exchange {
    int exId;
    string exName;
    map(int, string) coins = {
        { 1, "Bitcoin" }, { 2, "Ethereum" }, { 3, "Ripple" }
    };

    Exchange(int exId, string exName) {
        this.exId = exId;
        this.exName = exName;
    }

    func listNewCoin(int identifier, string name) {
        coins.add(identifier, name);
    }
}
```

#### Constructor parameters

constructor parameter는 생략가능하고, ASCL에서 지원하는 모든 타입을 사용할 수 있다.

```
Exchange() ...
Exchange(int identifier) ...

type exInfo struct {
    int identifier;
    string name;
    string address;
}
Exchange(exInfo info) ...
```

또한, array를 사용할 수도 있는데, undimensional array도 사용할 수 있다. 만약 array의 크기가 정의된 경우엔 반드시 같은 타입, 크기를 같는 array만 전달할 수 있고, 그렇지 않은 경우엔 타입이 같은 임의의 크기의 array를 전달할 수 있다.

```
Exchange(int ids[10]) ...       // allowed array having 10 elements
Exchange(string names[]) ...    // allowed array having any elements
```

#### Constructor returns

constructor는 return type을 가질 수 없다.

### Functions

function은 일련의 task를 수행하기 위한 기본 단위다.

<pre>
<a name="func_decl">FunctionDecl</a>   = <a href="#func_spec">FunctionSpec</a> <a href="#func_body">FunctionBody</a> ;
<a name="func_sepc">FunctionSpec</a>   = <a href="#mod">modifier</a> "func" <a href="#identifier">identifier</a> "(" [ <a href="#param_list">ParameterList</a> ] ")" [ <a href="#type">Type</a> ] ;
<a name="func_body">FunctionBody</a>   = <a href="#block_stmt">BlockStmt</a> ;
</pre>

#### Function specifications

function specification은 modifier와 이름, parameter list, return type list로 구성된다.

##### Function modifiers

function modifier는 각 function의 속성을 정의하는 것으로 다음과 같다.

<pre>
<a name="mod">modifier</a> = [ "public" ] [ "payable" ] ;
</pre>

function도 variable과 마찬가지로 기본적으로 private 속성을 가지나, public을 정의한 경우 외부 smart contract에서 참조할 수 있다.

> **TODO** 구현 후 보완 필요

payable modifier는 token 전송을 위한 필수 modifier다. 이 modifer가 정의되지 않은 경우엔 모든 token 송수신이 거절된다.

>readonly modifier는 smart contract의 상태를 변경시키지 않는 함수를 정의한다. 다음은 상태를 변경시키는 행위의 유형이다.
>
>* state variable에 값을 할당한다.
>* external smart contract를 참조한다.
>* payable modifier가 있는 함수를 호출한다.
>* readonly modifier가 없는 함수를 호출한다.

```
func f1 ...
public func f2 ...
payable func f3 ...
public payable f4 ...
```

##### Function name

function의 이름은 하나의 smart contract 내에서 unique해야 하고, 일반적인 <a href="#identifier">identifier</a> 규칙을 따른다.

```
func f ...
func get_id ...
func getName2 ...
```

##### Function parameters

function parameter는 constructor parameter와 동일하다.

##### Function returns

function은 1개 이하의 value를 return할 수 있는데, 이때 반환하는 value의 타입을 function specification에 미리 정의해야 하며, ASCL에서 지원하는 모든 타입을 사용할 수 있다.

```
func f1() {
    return;
}

func f2() int {
    return 1;
}
```

만약 return type이 정의되어 있으나, return statement가 없는 경우엔 에러가 발생한다.

##### Function overloading

> **TODO** 구현 후 설명 필요

#### Built-in functions

> **TODO** 구현 후 설명 필요

## SQL extensions

> **TODO** 구현 후 설명 필요

### DML, Query expressions

<pre>
<a name="sql_exp">SqlExp</a> = ? insert statement ? |
         ? update statement ? |
         ? delete statement ? |
         ? select statement ? ;
</pre>

### DDL statements

<pre>
<a name="ddl_stmt">DdlStmt</a> = ? create index ... ? |
          ? create table ... ? |
          ? drop index ... ? |
          ? drop table ... ? ;
</pre>

## Blockchain extensions

> **TODO** 구현 후 설명 필요

