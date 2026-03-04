package tests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/weaweawe01/velocity-ast/internal/dump"
	"github.com/weaweawe01/velocity-ast/internal/parser"
)

func TestParse1(t *testing.T) {
	tpl := "#set($x='') #set($rt=$x.class.forName('java.lang.Runtime')) #set($chr=$x.class.forName('java.lang.Character')) #set($str=$x.class.forName('java.lang.String')) #set($ex=$rt.getRuntime().exec('whoami')) $ex.waitFor() #set($out=$ex.getInputStream()) #foreach($i in [1..$out.available()])$str.valueOf($chr.toChars($out.read()))#end"
	root, tokens, err := parser.Parse(tpl)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	token := dump.Render(root, tokens)
	ast := `ASTprocess [id=0, info=0, invalid=false, tokens=[#set(], [$x], [=], [''], [)], [ ], [#set(], [$rt]...] -> #set(
├── ASTSetDirective [id=26, info=0, invalid=false, tokens=[#set(], [$x], [=], [''], [)]] -> #set(
│   ├── ASTReference [id=20, info=0, invalid=false, tokens=[$x]] -> $x
│   └── ASTExpression [id=27, info=0, invalid=false, tokens=['']] -> ''
│       └── ASTStringLiteral [id=9, info=0, invalid=false, tokens=['']] -> ''
├── ASTText [id=2, info=0, invalid=false, tokens=[ ]] ->  
├── ASTSetDirective [id=26, info=0, invalid=false, tokens=[#set(], [$rt], [=], [$x], [.], [class], [.], [for...] -> #set(
│   ├── ASTReference [id=20, info=0, invalid=false, tokens=[$rt]] -> $rt
│   └── ASTExpression [id=27, info=0, invalid=false, tokens=[$x], [.], [class], [.], [forName], [(], ['java.la...] -> $x
│       └── ASTReference [id=20, info=0, invalid=false, tokens=[$x], [.], [class], [.], [forName], [(], ['java.la...] -> $x
│           ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[class]] -> class
│           └── ASTMethod [id=18, info=0, invalid=false, tokens=[forName], [(], ['java.lang.Runtime'], [)]] -> forName
│               ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[forName]] -> forName
│               └── ASTExpression [id=27, info=0, invalid=false, tokens=['java.lang.Runtime']] -> 'java.lang.Runtime'
│                   └── ASTStringLiteral [id=9, info=0, invalid=false, tokens=['java.lang.Runtime']] -> 'java.lang.Runtime'
├── ASTText [id=2, info=0, invalid=false, tokens=[ ]] ->  
├── ASTSetDirective [id=26, info=0, invalid=false, tokens=[#set(], [$chr], [=], [$x], [.], [class], [.], [fo...] -> #set(
│   ├── ASTReference [id=20, info=0, invalid=false, tokens=[$chr]] -> $chr
│   └── ASTExpression [id=27, info=0, invalid=false, tokens=[$x], [.], [class], [.], [forName], [(], ['java.la...] -> $x
│       └── ASTReference [id=20, info=0, invalid=false, tokens=[$x], [.], [class], [.], [forName], [(], ['java.la...] -> $x
│           ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[class]] -> class
│           └── ASTMethod [id=18, info=0, invalid=false, tokens=[forName], [(], ['java.lang.Character'], [)]] -> forName
│               ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[forName]] -> forName
│               └── ASTExpression [id=27, info=0, invalid=false, tokens=['java.lang.Character']] -> 'java.lang.Character'
│                   └── ASTStringLiteral [id=9, info=0, invalid=false, tokens=['java.lang.Character']] -> 'java.lang.Character'
├── ASTText [id=2, info=0, invalid=false, tokens=[ ]] ->  
├── ASTSetDirective [id=26, info=0, invalid=false, tokens=[#set(], [$str], [=], [$x], [.], [class], [.], [fo...] -> #set(
│   ├── ASTReference [id=20, info=0, invalid=false, tokens=[$str]] -> $str
│   └── ASTExpression [id=27, info=0, invalid=false, tokens=[$x], [.], [class], [.], [forName], [(], ['java.la...] -> $x
│       └── ASTReference [id=20, info=0, invalid=false, tokens=[$x], [.], [class], [.], [forName], [(], ['java.la...] -> $x
│           ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[class]] -> class
│           └── ASTMethod [id=18, info=0, invalid=false, tokens=[forName], [(], ['java.lang.String'], [)]] -> forName
│               ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[forName]] -> forName
│               └── ASTExpression [id=27, info=0, invalid=false, tokens=['java.lang.String']] -> 'java.lang.String'
│                   └── ASTStringLiteral [id=9, info=0, invalid=false, tokens=['java.lang.String']] -> 'java.lang.String'
├── ASTText [id=2, info=0, invalid=false, tokens=[ ]] ->  
├── ASTSetDirective [id=26, info=0, invalid=false, tokens=[#set(], [$ex], [=], [$rt], [.], [getRuntime], [(]...] -> #set(
│   ├── ASTReference [id=20, info=0, invalid=false, tokens=[$ex]] -> $ex
│   └── ASTExpression [id=27, info=0, invalid=false, tokens=[$rt], [.], [getRuntime], [(], [)], [.], [exec], [...] -> $rt
│       └── ASTReference [id=20, info=0, invalid=false, tokens=[$rt], [.], [getRuntime], [(], [)], [.], [exec], [...] -> $rt
│           ├── ASTMethod [id=18, info=0, invalid=false, tokens=[getRuntime], [(], [)]] -> getRuntime
│           │   └── ASTIdentifier [id=10, info=0, invalid=false, tokens=[getRuntime]] -> getRuntime
│           └── ASTMethod [id=18, info=0, invalid=false, tokens=[exec], [(], ['whoami'], [)]] -> exec
│               ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[exec]] -> exec
│               └── ASTExpression [id=27, info=0, invalid=false, tokens=['whoami']] -> 'whoami'
│                   └── ASTStringLiteral [id=9, info=0, invalid=false, tokens=['whoami']] -> 'whoami'
├── ASTText [id=2, info=0, invalid=false, tokens=[ ]] ->  
├── ASTReference [id=20, info=0, invalid=false, tokens=[$ex], [.], [waitFor], [(], [)]] -> $ex
│   └── ASTMethod [id=18, info=0, invalid=false, tokens=[waitFor], [(], [)]] -> waitFor
│       └── ASTIdentifier [id=10, info=0, invalid=false, tokens=[waitFor]] -> waitFor
├── ASTText [id=2, info=0, invalid=false, tokens=[ ]] ->   
├── ASTSetDirective [id=26, info=0, invalid=false, tokens=[#set(], [$out], [=], [$ex], [.], [getInputStream]...] -> #set(
│   ├── ASTReference [id=20, info=0, invalid=false, tokens=[$out]] -> $out
│   └── ASTExpression [id=27, info=0, invalid=false, tokens=[$ex], [.], [getInputStream], [(], [)]] -> $ex
│       └── ASTReference [id=20, info=0, invalid=false, tokens=[$ex], [.], [getInputStream], [(], [)]] -> $ex
│           └── ASTMethod [id=18, info=0, invalid=false, tokens=[getInputStream], [(], [)]] -> getInputStream
│               └── ASTIdentifier [id=10, info=0, invalid=false, tokens=[getInputStream]] -> getInputStream
├── ASTText [id=2, info=0, invalid=false, tokens=[ ]] ->  
└── ASTDirective [ASTDirective [id=13, info=0, invalid=false, tokens=[#foreach], [(], [$i], [ ], [in], [ ], [[], [1], [...], directiveName=foreach] -> #foreach
    ├── ASTReference [id=20, info=0, invalid=false, tokens=[$i]] -> $i
    ├── ASTWord [id=11, info=0, invalid=false, tokens=[in]] -> in
    ├── ASTIntegerRange [id=17, info=0, invalid=false, tokens=[[], [1], [..], [$out], [.], [available], [(], [)]...] -> [
    │   ├── ASTIntegerLiteral [id=8, info=0, invalid=false, tokens=[1]] -> 1
    │   └── ASTReference [id=20, info=0, invalid=false, tokens=[$out], [.], [available], [(], [)]] -> $out
    │       └── ASTMethod [id=18, info=0, invalid=false, tokens=[available], [(], [)]] -> available
    │           └── ASTIdentifier [id=10, info=0, invalid=false, tokens=[available]] -> available
    └── ASTBlock [id=14, info=0, invalid=false, tokens=[$str], [.], [valueOf], [(], [$chr], [.], [toChars...] -> $str
        └── ASTReference [id=20, info=0, invalid=false, tokens=[$str], [.], [valueOf], [(], [$chr], [.], [toChars...] -> $str
            └── ASTMethod [id=18, info=0, invalid=false, tokens=[valueOf], [(], [$chr], [.], [toChars], [(], [$out...] -> valueOf
                ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[valueOf]] -> valueOf
                └── ASTExpression [id=27, info=0, invalid=false, tokens=[$chr], [.], [toChars], [(], [$out], [.], [read], ...] -> $chr
                    └── ASTReference [id=20, info=0, invalid=false, tokens=[$chr], [.], [toChars], [(], [$out], [.], [read], ...] -> $chr
                        └── ASTMethod [id=18, info=0, invalid=false, tokens=[toChars], [(], [$out], [.], [read], [(], [)], [)]] -> toChars
                            ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[toChars]] -> toChars
                            └── ASTExpression [id=27, info=0, invalid=false, tokens=[$out], [.], [read], [(], [)]] -> $out
                                └── ASTReference [id=20, info=0, invalid=false, tokens=[$out], [.], [read], [(], [)]] -> $out
                                    └── ASTMethod [id=18, info=0, invalid=false, tokens=[read], [(], [)]] -> read
                                        └── ASTIdentifier [id=10, info=0, invalid=false, tokens=[read]] -> read`
	if strings.TrimSpace(token) == strings.TrimSpace(ast) {
	} else {
		fmt.Println(token)
		t.Fatalf("对比失败")
	}

}

func TestParse2(t *testing.T) {
	tpl := "#set($e=666);$e.getClass().forName(\"java.lang.Runtime\").getMethod(\"getRuntime\",null).invoke(null,null).exec(\"calc\")"
	root, tokens, err := parser.Parse(tpl)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	token := dump.Render(root, tokens)

	ast := `ASTprocess [id=0, info=0, invalid=false, tokens=[#set(], [$e], [=], [666], [)], [;], [$e], [.], [g...] -> #set(
├── ASTSetDirective [id=26, info=0, invalid=false, tokens=[#set(], [$e], [=], [666], [)]] -> #set(
│   ├── ASTReference [id=20, info=0, invalid=false, tokens=[$e]] -> $e
│   └── ASTExpression [id=27, info=0, invalid=false, tokens=[666]] -> 666
│       └── ASTIntegerLiteral [id=8, info=0, invalid=false, tokens=[666]] -> 666
├── ASTText [id=2, info=0, invalid=false, tokens=[;]] -> ;
└── ASTReference [id=20, info=0, invalid=false, tokens=[$e], [.], [getClass], [(], [)], [.], [forName], [...] -> $e
    ├── ASTMethod [id=18, info=0, invalid=false, tokens=[getClass], [(], [)]] -> getClass
    │   └── ASTIdentifier [id=10, info=0, invalid=false, tokens=[getClass]] -> getClass
    ├── ASTMethod [id=18, info=0, invalid=false, tokens=[forName], [(], ["java.lang.Runtime"], [)]] -> forName
    │   ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[forName]] -> forName
    │   └── ASTExpression [id=27, info=0, invalid=false, tokens=["java.lang.Runtime"]] -> "java.lang.Runtime"
    │       └── ASTStringLiteral [id=9, info=0, invalid=false, tokens=["java.lang.Runtime"]] -> "java.lang.Runtime"
    ├── ASTMethod [id=18, info=0, invalid=false, tokens=[getMethod], [(], ["getRuntime"], [,], [null], [)]] -> getMethod
    │   ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[getMethod]] -> getMethod
    │   ├── ASTExpression [id=27, info=0, invalid=false, tokens=["getRuntime"]] -> "getRuntime"
    │   │   └── ASTStringLiteral [id=9, info=0, invalid=false, tokens=["getRuntime"]] -> "getRuntime"
    │   └── ASTExpression [id=27, info=0, invalid=false, tokens=[null]] -> null
    │       └── ASTReference [id=20, info=0, invalid=false, tokens=[null]] -> null
    ├── ASTMethod [id=18, info=0, invalid=false, tokens=[invoke], [(], [null], [,], [null], [)]] -> invoke
    │   ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[invoke]] -> invoke
    │   ├── ASTExpression [id=27, info=0, invalid=false, tokens=[null]] -> null
    │   │   └── ASTReference [id=20, info=0, invalid=false, tokens=[null]] -> null
    │   └── ASTExpression [id=27, info=0, invalid=false, tokens=[null]] -> null
    │       └── ASTReference [id=20, info=0, invalid=false, tokens=[null]] -> null
    └── ASTMethod [id=18, info=0, invalid=false, tokens=[exec], [(], ["calc"], [)]] -> exec
        ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[exec]] -> exec
        └── ASTExpression [id=27, info=0, invalid=false, tokens=["calc"]] -> "calc"
            └── ASTStringLiteral [id=9, info=0, invalid=false, tokens=["calc"]] -> "calc"
`
	if strings.TrimSpace(token) == strings.TrimSpace(ast) {
	} else {
		fmt.Println(token)
		t.Fatalf("对比失败")
	}

}

func ParseCheck(name string, t *testing.T) {
	file := fmt.Sprintf("../testdata/cases/%s.vtl", name)
	tplBytes, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read cases file: %v", err)
	}
	astFile := fmt.Sprintf("../testdata/expected-java/%s.ast", name)
	root, tokens, err := parser.Parse(string(tplBytes))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	token := dump.Render(root, tokens)
	// 读取文件testdata/expected-java/braced_if.ast的内容到expected中
	expectedBytes, err := os.ReadFile(astFile)
	if err != nil {
		t.Fatalf("read expected file: %v", err)
	}
	expected := string(expectedBytes)

	// token 一行行对比 expected
	tokenLines := strings.Split(strings.TrimSpace(token), "\n")
	expectedLines := strings.Split(strings.TrimSpace(expected), "\n")
	if len(tokenLines) != len(expectedLines) {
		fmt.Printf("line count mismatch: got %d lines, expected %d lines\n", len(tokenLines), len(expectedLines))
		fmt.Println("Got:")
		fmt.Println(token)
		fmt.Println("Expected:")
		fmt.Println(expected)
		t.Fatalf("line count mismatch")
	}

	// 一行行对比
	for i := 0; i < len(tokenLines); i++ {
		if strings.TrimSpace(tokenLines[i]) != strings.TrimSpace(expectedLines[i]) {
			fmt.Printf("line %d mismatch:\nGot: %s\nExpected: %s\n", i+1, tokenLines[i], expectedLines[i])
			t.Fatalf("line %d mismatch", i+1)
		}
	}

}

func TestParseBracedIf(t *testing.T) {
	ParseCheck("braced_if", t)
}

func TestParseBreakDirective(t *testing.T) {
	ParseCheck("break_directive", t)
}

func TestParseBreakNewline(t *testing.T) {
	ParseCheck("break_newline", t)
}

func TestParseBreakThenText(t *testing.T) {
	ParseCheck("break_then_text", t)
}

func TestParseCommentBlock(t *testing.T) {
	ParseCheck("comment_block", t)
}

func TestParseCommentLine(t *testing.T) {
	ParseCheck("comment_line", t)
}

func TestParseDefineBasic(t *testing.T) {
	ParseCheck("define_basic", t)
}

func TestParseEscapeDouble(t *testing.T) {
	ParseCheck("escape_double", t)
}

func TestParseEvaluateRef(t *testing.T) {
	ParseCheck("evaluate_ref", t)
}

func TestParseExploitChain(t *testing.T) {
	ParseCheck("exploit_chain", t)
}

func TestParseForeachBasic(t *testing.T) {
	ParseCheck("foreach_basic", t)
}

func TestParseForeachElse(t *testing.T) {
	ParseCheck("foreach_else", t)
}

func TestParseIfAnd(t *testing.T) {
	ParseCheck("if_and", t)
}

func TestParseIfBasic(t *testing.T) {
	ParseCheck("if_basic", t)
}

func TestParseIfElse(t *testing.T) {
	ParseCheck("if_else", t)
}

func TestParseIfElseifElse(t *testing.T) {
	ParseCheck("if_elseif_else", t)
}

func TestParseIfEq(t *testing.T) {
	ParseCheck("if_eq", t)
}

func TestParseIfGe(t *testing.T) {
	ParseCheck("if_ge", t)
}

func TestParseIfGt(t *testing.T) {
	ParseCheck("if_gt", t)
}

func TestParseIfKeywordOps(t *testing.T) {
	ParseCheck("if_keyword_ops", t)
}

func TestParseIfLe(t *testing.T) {
	ParseCheck("if_le", t)
}

func TestParseIfLt(t *testing.T) {
	ParseCheck("if_lt", t)
}

func TestParseIfNe(t *testing.T) {
	ParseCheck("if_ne", t)
}

func TestParseIfNewlineBody(t *testing.T) {
	ParseCheck("if_newline_body", t)
}

func TestParseIfNot(t *testing.T) {
	ParseCheck("if_not", t)
}

func TestParseIfNotParenEq(t *testing.T) {
	ParseCheck("if_not_paren_eq", t)
}

func TestParseIfOr(t *testing.T) {
	ParseCheck("if_or", t)
}

func TestParseIfTrueFalse(t *testing.T) {
	ParseCheck("if_true_false", t)
}

func TestParseIncludeMulti(t *testing.T) {
	ParseCheck("include_multi", t)
}

func TestParseIncludeThenText(t *testing.T) {
	ParseCheck("include_then_text", t)
}

func TestParseMacroCallNoParen(t *testing.T) {
	ParseCheck("macro_call_no_paren", t)
}

func TestParseMacroDefault(t *testing.T) {
	ParseCheck("macro_default", t)
}

func TestParseMacroDefineCall(t *testing.T) {
	ParseCheck("macro_define_call", t)
}

func TestParseParseBasic(t *testing.T) {
	ParseCheck("parse_basic", t)
}

func TestParseParseNewline(t *testing.T) {
	ParseCheck("parse_newline", t)
}

func TestParseParseThenText(t *testing.T) {
	ParseCheck("parse_then_text", t)
}

func TestParseQuietRef(t *testing.T) {
	ParseCheck("quiet_ref", t)
}

func TestParseSetAdd(t *testing.T) {
	ParseCheck("set_add", t)
}

func TestParseSetArray(t *testing.T) {
	ParseCheck("set_array", t)
}

func TestParseSetBrace(t *testing.T) {
	ParseCheck("set_brace", t)
}

func TestParseSetDiv(t *testing.T) {
	ParseCheck("set_div", t)
}

func TestParseSetFloat(t *testing.T) {
	ParseCheck("set_float", t)
}

func TestParseSetFormalPropRef(t *testing.T) {
	ParseCheck("set_formal_prop_ref", t)
}

func TestParseSetFormalRef(t *testing.T) {
	ParseCheck("set_formal_ref", t)
}

func TestParseSetFormalRefAlt(t *testing.T) {
	ParseCheck("set_formal_ref_alt", t)
}

func TestParseSetIndexInt(t *testing.T) {
	ParseCheck("set_index_int", t)
}

func TestParseSetIndexRef(t *testing.T) {
	ParseCheck("set_index_ref", t)
}

func TestParseSetMap(t *testing.T) {
	ParseCheck("set_map", t)
}

func TestParseSetMul(t *testing.T) {
	ParseCheck("set_mul", t)
}

func TestParseSetNegFloat(t *testing.T) {
	ParseCheck("set_neg_float", t)
}

func TestParseSetNegRef(t *testing.T) {
	ParseCheck("set_neg_ref", t)
}

func TestParseSetParenMul(t *testing.T) {
	ParseCheck("set_paren_mul", t)
}

func TestParseSetPropRef(t *testing.T) {
	ParseCheck("set_prop_ref", t)
}

func TestParseSetRange(t *testing.T) {
	ParseCheck("set_range", t)
}

func TestParseSetSpace(t *testing.T) {
	ParseCheck("set_space", t)
}

func TestParseSetSub(t *testing.T) {
	ParseCheck("set_sub", t)
}

func TestParseStopDirective(t *testing.T) {
	ParseCheck("stop_directive", t)
}

func TestParseStopNewline(t *testing.T) {
	ParseCheck("stop_newline", t)
}

func TestParseStopThenText(t *testing.T) {
	ParseCheck("stop_then_text", t)
}

func TestParseTextblock(t *testing.T) {
	ParseCheck("textblock", t)
}
func TestParseRce1(t *testing.T) {
	ParseCheck("rce1", t)
}

func TestParseRce2(t *testing.T) {
	ParseCheck("rce2", t)
}
