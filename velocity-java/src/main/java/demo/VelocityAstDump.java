package demo;

import org.apache.velocity.Template;
import org.apache.velocity.runtime.RuntimeSingleton;
import org.apache.velocity.runtime.parser.Token;
import org.apache.velocity.runtime.parser.node.SimpleNode;

import java.io.BufferedReader;
import java.io.FileInputStream;
import java.io.InputStreamReader;
import java.io.StringReader;
import java.util.Properties;

public class VelocityAstDump
{
    private static final String DEFAULT_TEMPLATE ="#set($x='') #set($rt=$x.class.forName('java.lang.Runtime')) #set($chr=$x.class.forName('java.lang.Character')) #set($str=$x.class.forName('java.lang.String')) #set($ex=$rt.getRuntime().exec('id')) $ex.waitFor() #set($out=$ex.getInputStream()) #foreach($i in [1..$out.available()])$str.valueOf($chr.toChars($out.read()))#end";
    private static final String BRANCH = "├── ";
    private static final String LAST = "└── ";
    private static final String PIPE = "│   ";
    private static final String SPACE = "    ";

    public static void main(String[] args) throws Exception
    {
        String template = resolveTemplate(args);

        Properties properties = new Properties();
        RuntimeSingleton.init(properties);

        Template inlineTemplate = new Template();
        inlineTemplate.setName("inline-demo");

        SimpleNode ast = RuntimeSingleton.parse(
            new BufferedReader(new StringReader(template)),
            inlineTemplate
        );

        printNode(ast, "", true, true);
    }

    private static String resolveTemplate(String[] args) throws Exception
    {
        if (args == null || args.length == 0)
        {
            return DEFAULT_TEMPLATE;
        }

        String joined = String.join(" ", args).trim();
        joined = stripMatchingQuotes(joined);
        if (joined.startsWith("-e "))
        {
            return joined.substring(3);
        }
        if (joined.startsWith("-f "))
        {
            return readFileTemplate(joined.substring(3));
        }

        if (args.length == 1 && args[0] != null)
        {
            String one = args[0];
            one = stripMatchingQuotes(one);
            if (one.startsWith("-e "))
            {
                return one.substring(3);
            }
            if (one.startsWith("-f "))
            {
                return readFileTemplate(one.substring(3));
            }
        }

        if (args.length >= 2 && "-e".equals(args[0]))
        {
            return args[1];
        }

        if (args.length >= 2 && "-f".equals(args[0]))
        {
            return readFileTemplate(args[1]);
        }

        return String.join(" ", args);
    }

    private static String stripMatchingQuotes(String input)
    {
        if (input == null || input.length() < 2)
        {
            return input;
        }
        if ((input.startsWith("\"") && input.endsWith("\""))
            || (input.startsWith("'") && input.endsWith("'")))
        {
            return input.substring(1, input.length() - 1).trim();
        }
        return input;
    }

    private static String readFileTemplate(String path) throws Exception
    {
        StringBuilder sb = new StringBuilder();
        try (BufferedReader br = new BufferedReader(new InputStreamReader(new FileInputStream(path), "UTF-8")))
        {
            String line;
            boolean first = true;
            while ((line = br.readLine()) != null)
            {
                if (!first)
                {
                    sb.append('\n');
                }
                sb.append(line);
                first = false;
            }
        }
        return sb.toString();
    }

    private static void printNode(SimpleNode node, String prefix, boolean tail, boolean root)
    {
        String linePrefix = root ? "" : prefix + (tail ? LAST : BRANCH);
        System.out.println(linePrefix + formatVerboseNode(node));

        int childCount = node.jjtGetNumChildren();
        String childPrefix = root ? "" : prefix + (tail ? SPACE : PIPE);
        for (int i = 0; i < childCount; i++)
        {
            boolean childTail = (i == childCount - 1);
            printNode((SimpleNode) node.jjtGetChild(i), childPrefix, childTail, false);
        }
    }

    private static String formatVerboseNode(SimpleNode node)
    {
        String tokens = "";
        Token token = node.getFirstToken();
        if (token != null && token.image != null)
        {
            String special = "";
            if (token.specialToken != null && token.specialToken.image != null && node.getParser() != null)
            {
                String lineComment = node.getParser().lineComment();
                if (lineComment == null || !token.specialToken.image.startsWith(lineComment))
                {
                    special = token.specialToken.image;
                }
            }
            tokens = " -> " + special + token.image;
        }
        return node + tokens;
    }
}
