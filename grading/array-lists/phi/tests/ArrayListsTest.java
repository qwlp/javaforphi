package phi.tests;
import static org.junit.Assert.*;
import org.junit.Test;
import java.lang.reflect.*;
import java.util.*;
import lib.*;

public class ArrayListsTest {
    @Test public void stringListTransformsSuppliedWordsInBothCases() throws Exception {
        String[] words={"Apple","bANana","Kiwi"};
        String output=ConsoleSupport.run("main.StringListDemo","",words);
        for(String word:words) {
            assertTrue("Print uppercase form of every supplied word",output.contains(word.toUpperCase(Locale.ROOT)));
            assertTrue("Print lowercase form of every supplied word",output.contains(word.toLowerCase(Locale.ROOT)));
        }
    }
    @Test public void nameListReadsFourNamesInsteadOfHardcodingThem() throws Exception {
        String[] names={"Ada Lovelace","Alan Turing","Grace Hopper","Linus Torvalds"};
        String input="";for(String name:names) input+=name.replace(' ','\n')+"\n";
        String output=ConsoleSupport.run("main.NameListDemo",input);
        for(String name:names) assertTrue("Print each supplied full name: "+name,output.contains(name));
    }
    @Test public void nameMembershipUsesValueEquality() {
        List<Name> names=new ArrayList<>(); names.add(new Name("Ada","Lovelace"));
        assertTrue("contains must recognize a distinct Name with the same value",names.contains(new Name("Ada","Lovelace")));
        assertFalse(names.contains(new Name("Grace","Lovelace")));
    }
    private double invoke(String method,List<OrderLine> lines) throws Exception {
        Class<?> demo;
        try { demo=Class.forName("main.OrderListDemo"); } catch(ClassNotFoundException missing) { throw new AssertionError("Create main.OrderListDemo with totalCost(List<OrderLine>) and averageCost(List<OrderLine>)",missing); }
        Method operation;
        try { operation=demo.getMethod(method,List.class); } catch(NoSuchMethodException missing) { throw new AssertionError("Add public static "+method+"(List<OrderLine>) to OrderListDemo",missing); }
        return ((Number)operation.invoke(null,lines)).doubleValue();
    }
    @Test public void orderTotalsAndAveragesUseCurrentCollectionContents() throws Exception {
        List<OrderLine> lines=new ArrayList<>();
        assertEquals(0,invoke("totalCost",lines),0.0001);
        assertEquals("Empty average is zero",0,invoke("averageCost",lines),0.0001);
        Random random=new Random(4129);
        double total=0;
        for(int i=0;i<12;i++) {
            int price=1+random.nextInt(250),quantity=random.nextInt(8);
            lines.add(new OrderLine("item-"+i,price,quantity));total+=price*quantity;
            assertEquals("Total must include every order line",total,invoke("totalCost",lines),0.0001);
            assertEquals("Average must divide by current number of lines",total/lines.size(),invoke("averageCost",lines),0.0001);
        }
    }
    private static class FixedDie extends Die {
        int rolls; final int score;
        FixedDie(int score) { this.score=score; }
        public int getScore(){return score;} public void roll(){rolls++;}
    }
    @Test public void multipleDiceDelegatesRollsAndScoresAndCollectionOperations() {
        MultipleDice dice=new MultipleDice();
        assertTrue(dice.isEmpty());assertEquals(0,dice.getScore());
        FixedDie first=new FixedDie(3),second=new FixedDie(5);first.rolls=second.rolls=0;
        dice.addDie(first);dice.addDie(second);dice.roll();
        assertEquals(1,first.rolls);assertEquals(1,second.rolls);assertEquals(8,dice.getScore());assertSame(first,dice.getDie(0));
        dice.removeDie(0);assertEquals(1,dice.getSize());assertEquals(5,dice.getScore());
        dice.clear();assertTrue(dice.isEmpty());assertEquals(0,dice.getScore());
    }
}
