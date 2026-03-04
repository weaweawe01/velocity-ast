package demo;

import org.apache.velocity.Template;
import org.apache.velocity.runtime.RuntimeSingleton;
import org.apache.velocity.runtime.parser.Token;
import org.apache.velocity.runtime.parser.node.Node;
import org.apache.velocity.runtime.parser.node.SimpleNode;

import java.io.BufferedReader;
import java.io.StringReader;
import java.util.Properties;

/**
 * Velocity AST 树形可视化 Demo
 * <p>
 * 以缩进树形结构输出 Velocity 模板解析后的 AST 节点，
 * 包括节点类型、Token 内容及行列位置信息。
 * <p>
 * 用法:
 *   mvn compile exec:java -Dexec.mainClass=demo.VelocityAstTreeDemo
 *   mvn compile exec:java -Dexec.mainClass=demo.VelocityAstTreeDemo -Dexec.args="'#foreach(\$item in \$list)\$item #end'"
 */
public class VelocityAstTreeDemo
{
    /** 树形连接符 */
    private static final String BRANCH   = "├── ";
    private static final String LAST     = "└── ";
    private static final String PIPE     = "│   ";
    private static final String SPACE    = "    ";

    /** ANSI 颜色（终端支持时生效） */
    private static final String CYAN     = "\u001B[36m";
    private static final String YELLOW   = "\u001B[33m";
    private static final String GREEN    = "\u001B[32m";
    private static final String GRAY     = "\u001B[90m";
    private static final String RESET    = "\u001B[0m";

    private static final String[] SAMPLE_TEMPLATES = {
            "#set($e=666);$e.getClass().forName(\"java.lang.Runtime\").getMethod(\"getRuntime\",null).invoke(null,null).exec(\"calc\")"
    };

    public static void main(String[] args) throws Exception
    {
        // 初始化 Velocity 运行时
        Properties properties = new Properties();
        RuntimeSingleton.init(properties);

        if (args.length > 0)
        {
            // 用户自定义模板
            String template = String.join(" ", args);
            dumpAst("自定义模板", template);
        }
        else
        {
            // 展示所有内置示例
            for (int i = 0; i < SAMPLE_TEMPLATES.length; i++)
            {
                dumpAst("示例 " + (i + 1), SAMPLE_TEMPLATES[i]);
                if (i < SAMPLE_TEMPLATES.length - 1)
                {
                    System.out.println();
                }
            }
        }
    }

    /**
     * 解析模板并输出 AST 树
     */
    private static void dumpAst(String title, String templateText) throws Exception
    {
        System.out.println("╔══════════════════════════════════════════════════════════════╗");
        System.out.printf( "║  %s%-56s%s║%n", CYAN, title, RESET);
        System.out.println("╠══════════════════════════════════════════════════════════════╣");

        // 显示模板内容（截断过长模板）
        String display = templateText.replace("\n", "\\n").replace("\r", "\\r");
        if (display.length() > 56)
        {
            display = display.substring(0, 53) + "...";
        }
        System.out.printf("║  %s%-56s%s║%n", YELLOW, display, RESET);
        System.out.println("╚══════════════════════════════════════════════════════════════╝");

        // 解析
        Template inlineTemplate = new Template();
        inlineTemplate.setName("demo-" + title);
        SimpleNode ast = RuntimeSingleton.parse(
                new BufferedReader(new StringReader(templateText)),
                inlineTemplate
        );

        // 输出树
        printTree(ast, "");
    }

    /**
     * 递归打印 AST 节点的树形结构
     *
     * @param node   当前节点
     * @param prefix 当前行前缀（用于对齐树形连线）
     */
    private static void printTree(Node node, String prefix)
    {
        SimpleNode simpleNode = (SimpleNode) node;
        int childCount = simpleNode.jjtGetNumChildren();

        // 根节点特殊处理
        if (prefix.isEmpty())
        {
            System.out.println(formatNode(simpleNode));
            for (int i = 0; i < childCount; i++)
            {
                boolean isLast = (i == childCount - 1);
                String connector = isLast ? LAST : BRANCH;
                String childPrefix = isLast ? SPACE : PIPE;

                System.out.print(prefix + connector);
                printNodeLine(simpleNode.jjtGetChild(i));
                printChildren(simpleNode.jjtGetChild(i), prefix + childPrefix);
            }
        }
    }

    /**
     * 打印当前节点行（不递归子节点）
     */
    private static void printNodeLine(Node node)
    {
        System.out.println(formatNode((SimpleNode) node));
    }

    /**
     * 递归打印子节点
     */
    private static void printChildren(Node node, String prefix)
    {
        SimpleNode simpleNode = (SimpleNode) node;
        int childCount = simpleNode.jjtGetNumChildren();

        for (int i = 0; i < childCount; i++)
        {
            boolean isLast = (i == childCount - 1);
            String connector = isLast ? LAST : BRANCH;
            String childPrefix = prefix + (isLast ? SPACE : PIPE);

            System.out.print(prefix + connector);
            System.out.println(formatNode((SimpleNode) simpleNode.jjtGetChild(i)));
            printChildren(simpleNode.jjtGetChild(i), childPrefix);
        }
    }

    /**
     * 格式化单个节点信息：类型名 + Token值 + 位置
     */
    private static String formatNode(SimpleNode node)
    {
        StringBuilder sb = new StringBuilder();

        // 节点类型
        String typeName = node.getClass().getSimpleName();
        sb.append(GREEN).append(typeName).append(RESET);

        // Token 值
        Token firstToken = node.getFirstToken();
        if (firstToken != null && firstToken.image != null && !firstToken.image.isEmpty())
        {
            String tokenStr = firstToken.image
                    .replace("\n", "\\n")
                    .replace("\r", "\\r")
                    .replace("\t", "\\t");

            if (tokenStr.length() > 30)
            {
                tokenStr = tokenStr.substring(0, 27) + "...";
            }
            sb.append(YELLOW).append(" \"").append(tokenStr).append("\"").append(RESET);
        }

        // 行列位置
        if (firstToken != null)
        {
            sb.append(GRAY)
                    .append(" [L").append(firstToken.beginLine)
                    .append(":C").append(firstToken.beginColumn)
                    .append("]")
                    .append(RESET);
        }

        return sb.toString();
    }
}
