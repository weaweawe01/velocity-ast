package demo;

import org.apache.velocity.Template;
import org.apache.velocity.runtime.RuntimeSingleton;
import org.apache.velocity.runtime.parser.Token;
import org.apache.velocity.runtime.parser.node.Node;
import org.apache.velocity.runtime.parser.node.SimpleNode;

import java.io.BufferedReader;
import java.io.StringReader;
import java.util.Properties;

public class VelocityAstDumpPretty
{
    private static final String DEFAULT_TEMPLATE =
        "#set($e=666);$e.getClass().forName(\"java.lang.Runtime\").getMethod(\"getRuntime\",null).invoke(null,null).exec(\"calc\")";

    public static void main(String[] args) throws Exception
    {
        String template = args.length > 0 ? String.join(" ", args) : DEFAULT_TEMPLATE;

        Properties properties = new Properties();
        RuntimeSingleton.init(properties);

        Template inlineTemplate = new Template();
        inlineTemplate.setName("inline-demo");

        SimpleNode ast = RuntimeSingleton.parse(
            new BufferedReader(new StringReader(template)),
            inlineTemplate
        );

        printAst(ast, 0);
    }

    private static void printAst(Node node, int depth)
    {
        SimpleNode simpleNode = (SimpleNode) node;
        Token firstToken = simpleNode.getFirstToken();
        String token = firstToken == null ? "" : firstToken.image;

        System.out.println(indent(depth) + simpleNode.getClass().getSimpleName() + formatToken(token));
        for (int i = 0; i < simpleNode.jjtGetNumChildren(); i++)
        {
            printAst(simpleNode.jjtGetChild(i), depth + 1);
        }
    }

    private static String indent(int depth)
    {
        StringBuilder sb = new StringBuilder(depth * 2);
        for (int i = 0; i < depth; i++)
        {
            sb.append("  ");
        }
        return sb.toString();
    }

    private static String formatToken(String token)
    {
        if (token == null || token.isEmpty())
        {
            return "";
        }

        String normalized = token.replace("\n", "\\n").replace("\r", "\\r");
        int maxLength = 40;
        if (normalized.length() > maxLength)
        {
            normalized = normalized.substring(0, maxLength) + "...";
        }
        return " -> " + normalized;
    }
}
