package phi.tests;
import static org.junit.Assert.*;
import org.junit.Test;
import java.util.*;
import java.util.regex.*;

public class BasicProgramsTest {
    @Test public void temperatureConversionRetainsFractionalDegrees() throws Exception {
        String output = ConsoleSupport.run("primitives.Converter", "");
        assertTrue("21 Celsius must convert to 69.8 Fahrenheit; avoid integer division.", output.matches("(?s).*69[.,]8(?:0*)\\b.*"));
    }
    @Test public void gradeLabelsRespectEveryBoundary() throws Exception {
        int[] marks = {0, 17, 39, 40, 46, 59, 60, 64, 69, 70, 83, 100};
        for (int mark : marks) {
            String expected = mark < 40 ? "Fail" : mark < 60 ? "Pass" : mark < 70 ? "Merit" : "Distinction";
            String output = ConsoleSupport.run("controlstructures.GradeMark", mark+"\n", ""+mark).trim();
            assertEquals("GradeMark must classify the supplied mark " + mark, expected, output);
        }
    }
    @Test public void allSevenDaysHaveCorrectNamesAndClassification() throws Exception {
        String[] names = {"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"};
        for (int day = 1; day <= 7; day++) {
            String output = ConsoleSupport.run("controlstructures.DaysOfWeek", day+"\n", ""+day);
            assertTrue("Incorrect name for day " + day, output.contains(names[day-1]));
            assertTrue("Incorrect weekday/weekend classification for day " + day, output.contains(day <= 5 ? "Weekday" : "Weekend"));
        }
    }
    @Test public void outOfRangeDaysAreUnknown() throws Exception {
        for (int day : new int[]{-1, 0, 8, 42}) assertTrue("Out-of-range days must print Unknown day", ConsoleSupport.run("controlstructures.DaysOfWeek", day+"\n", ""+day).contains("Unknown day"));
    }
    @Test public void multiplicationTablesContainAll144ProductsInOrder() throws Exception {
        String output = ConsoleSupport.run("controlstructures.TimesTable", "");
        Matcher numbers = Pattern.compile("\\d+").matcher(output);
        List<Integer> actual = new ArrayList<>();
        while (numbers.find()) actual.add(Integer.parseInt(numbers.group()));
        assertEquals("Print the 12 products for each of tables 1 through 12 (no numeric labels).", 144, actual.size());
        int index=0;
        for (int table=1; table<=12; table++) for (int factor=1; factor<=12; factor++) assertEquals("Incorrect product at table " + table + ", factor " + factor, table*factor, actual.get(index++).intValue());
    }
}
