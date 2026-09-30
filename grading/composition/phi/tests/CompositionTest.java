package phi.tests;
import static org.junit.Assert.*;
import org.junit.Test;
import lib.dice.*;
import lib.employee.*;

public class CompositionTest {
    private static class FixedDie extends Die {
        int rolls; final int score;
        FixedDie(int score){this.score=score;}
        public void roll(){rolls++;} public int getScore(){return score;}
    }
    @Test public void pairOfDicePreservesObjectsAndDelegatesBehavior() {
        for(int redScore=1;redScore<=6;redScore++) {
            FixedDie red=new FixedDie(redScore),blue=new FixedDie(2+redScore%3);
            red.rolls=blue.rolls=0;
            PairOfDice pair=new PairOfDice(red,blue);
            assertSame(red,pair.getRed());assertSame(blue,pair.getBlue());
            Rollable rollable=pair;rollable.roll();
            assertEquals("Roll must reach the red die",1,red.rolls);
            assertEquals("Roll must reach the blue die",1,blue.rolls);
            assertEquals("Score is the sum of the contained dice",redScore+2+redScore%3,rollable.getScore());
        }
    }
    @Test public void employeeUsesContainedNameAndDateAndUpdatesSalary() {
        Name name=new Name("Ada","Lovelace");Date date=new Date(10,12,2020);
        Employee employee=new Employee(name,date,12345.5);
        assertSame(name,employee.getName());assertSame(date,employee.getStartDate());assertEquals(12345.5,employee.getSalary(),0.0001);
        name.setFirstName("Grace");assertEquals("Grace",employee.getName().getFirstName());
        employee.setSalary(54321.25);assertEquals(54321.25,employee.getSalary(),0.0001);
        Name replacement=new Name("Alan","Turing");Date later=new Date(2,3,2021);
        employee.setName(replacement);employee.setStartDate(later);
        assertSame(replacement,employee.getName());assertSame(later,employee.getStartDate());
    }
    @Test public void employeeEqualityUsesContainedValuesAndSalary() {
        Employee first=new Employee(new Name("Ada","Lovelace"),new Date(1,2,2020),15000);
        Employee equal=new Employee(new Name("Ada","Lovelace"),new Date(1,2,2020),15000);
        assertEquals(first,equal);assertEquals(equal,first);
        equal.setSalary(15001);assertNotEquals(first,equal);assertNotEquals(first,null);assertNotEquals(first,"employee");
    }
}
