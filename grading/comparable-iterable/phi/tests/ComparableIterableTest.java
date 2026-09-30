package phi.tests;
import static org.junit.Assert.*;
import org.junit.Test;
import java.util.ArrayList;
import java.util.List;
import lib.employeeRegister.*;

public class ComparableIterableTest {
    @Test public void namesCompareFamilyNameBeforeFirstName() {
        assertTrue(new Name("Zoe","Adams").compareTo(new Name("Ada","Brown"))<0);
        assertTrue(new Name("Ada","Brown").compareTo(new Name("Zoe","Brown"))<0);
        assertEquals(0,new Name("Ada","Brown").compareTo(new Name("Ada","Brown")));
    }
    @Test public void datesCompareYearThenMonthThenDay() {
        assertTrue(new Date(31,12,2019).compareTo(new Date(1,1,2020))<0);
        assertTrue(new Date(31,1,2020).compareTo(new Date(1,2,2020))<0);
        assertTrue(new Date(1,2,2020).compareTo(new Date(2,2,2020))<0);
        assertEquals(0,new Date(1,2,2020).compareTo(new Date(1,2,2020)));
    }
    @Test public void employeeComparisonUsesDateAndSalaryAsTieBreakers() {
        Employee early=new Employee(new Name("Ada","Lovelace"),new Date(1,1,2020),90000);
        Employee later=new Employee(new Name("Ada","Lovelace"),new Date(2,1,2020),10000);
        assertTrue(early.compareTo(later)<0);
        Employee higher=new Employee(new Name("Ada","Lovelace"),new Date(2,1,2020),10001);
        assertTrue(later.compareTo(higher)<0);assertTrue(higher.compareTo(later)>0);
        assertEquals(0,later.compareTo(new Employee(new Name("Ada","Lovelace"),new Date(2,1,2020),10000)));
    }
    @Test public void registerIterationAndSortingReflectCurrentContents() {
        EmployeeRegister register=new EmployeeRegister();assertFalse(register.iterator().hasNext());
        Employee a=new Employee(new Name("Zoe","Brown"),new Date(1,1,2020),1);
        Employee b=new Employee(new Name("Ada","Adams"),new Date(1,1,2020),2);
        Employee c=new Employee(new Name("Ada","Brown"),new Date(1,1,2020),3);
        register.addEmployee(a);register.addEmployee(b);register.addEmployee(c);
        List<Employee> actual=new ArrayList<>();for(Employee employee:register)actual.add(employee);assertEquals(List.of(a,b,c),actual);
        register.sortEmployeeRegister();actual.clear();for(Employee employee:register)actual.add(employee);assertEquals(List.of(b,c,a),actual);
        register.removeEmployee(1);actual.clear();for(Employee employee:register)actual.add(employee);assertEquals(List.of(b,a),actual);
    }
}
