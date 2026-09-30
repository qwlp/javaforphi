package phi.tests;
import static org.junit.Assert.*;
import org.junit.Test;
import java.util.Locale;

public class UsingObjectsTest {
    @Test public void initialsAndEmailsUseTheSuppliedName() throws Exception {
        String[] names = {"David Beckham", "Ada Lovelace", "grace hopper", "Alan Turing"};
        for (String name : names) {
            String[] parts=name.split(" ");
            String initials=(""+parts[0].charAt(0)+parts[1].charAt(0)).toUpperCase(Locale.ROOT);
            String output=ConsoleSupport.run("strings.Initials",name+"\n",name);
            assertTrue("Print uppercase initials for the supplied name",output.contains(initials));
            assertTrue("Print lowercase initials@email.dmu.ac.uk for the supplied name",output.contains(initials.toLowerCase(Locale.ROOT)+"@email.dmu.ac.uk"));
        }
    }
    @Test public void sixFruitNamesArePrintedInUppercase() throws Exception {
        String[] words=ConsoleSupport.run("strings.StringArrayDemo","").trim().split("\\s+");
        assertEquals("Print exactly six fruit names separated by whitespace",6,words.length);
        for(String word:words) assertTrue("Fruit names must be uppercase",word.matches("[A-Z]+"));
    }
    @Test public void lowercaseReturnsACopyAndPreservesOriginal() throws Exception {
        for (String original : new String[]{"HeLLo", "JaVa", "PhiCourse"}) {
            String output=ConsoleSupport.run("strings.ImmutableDemo",original+"\n",original);
            assertTrue("Print the original String after calling toLowerCase",output.contains(original));
            assertTrue("Print the returned lowercase copy",output.contains(original.toLowerCase(Locale.ROOT)));
        }
    }
    @Test public void scannerConsumesNewlineAfterDecimalInput() throws Exception {
        String sentence="The next line stays intact";
        String output=ConsoleSupport.run("scanner.ScannerSurprise","7\nFirst sentence\n2.5\n"+sentence+"\n");
        assertTrue("After nextDouble, consume the newline and read the final sentence; print its text or its correct length.",output.contains(sentence) || output.matches("(?s).*sentence\\s+"+sentence.length()+"\\b.*"));
    }
}
