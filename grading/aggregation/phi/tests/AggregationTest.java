package phi.tests;
import static org.junit.Assert.*;
import org.junit.Test;
import lib.playlist.*;
import lib.employeeRegister.*;

public class AggregationTest {
    @Test public void playlistUpdatesSizeDurationAndSearchAfterMutations() {
        PlayList list=new PlayList("Study");assertTrue(list.isPlayListEmpty());assertEquals(0,list.getTotalTime());
        Song first=new Song("First",123,"Ada"),second=new Song("Second",241,"Grace");
        list.addSong(first);list.addSong(second);
        assertEquals(2,list.numberOfSongs());assertEquals(364,list.getTotalTime());assertSame(first,list.getSong(0));
        assertTrue(list.searchSongByTitle("First"));assertFalse(list.searchSongByTitle("Missing"));
        assertEquals("Duration threshold is strictly less than",1,list.countSongsWithDurationLessThan(241));
        list.removeSong(0);assertEquals(241,list.getTotalTime());assertFalse(list.searchSongByTitle("First"));
        list.clearPlayList();assertTrue(list.isPlayListEmpty());assertEquals(0,list.getTotalTime());
    }
    @Test public void playlistBoundaryOperationsDoNotThrowOrChangeContents() {
        PlayList list=new PlayList();Song song=new Song("Only",17,"Artist");list.addSong(song);
        for(int index:new int[]{-1,1,20}) { list.removeSong(index);list.moveUp(index);list.moveDown(index);assertEquals(1,list.numberOfSongs());assertSame(song,list.getSong(0)); }
        list.moveUp(0);list.moveDown(0);assertSame(song,list.getSong(0));
        assertNull(list.getSong(-1));assertNull(list.getSong(1));
    }
    @Test public void playlistMovesSongsOnePositionAndPreservesDuration() {
        PlayList list=new PlayList();Song a=new Song("A",10,"a"),b=new Song("B",20,"b"),c=new Song("C",30,"c");
        list.addSong(a);list.addSong(b);list.addSong(c);list.moveUp(2);
        assertSame(c,list.getSong(1));assertSame(b,list.getSong(2));list.moveDown(0);assertSame(a,list.getSong(1));assertEquals(60,list.getTotalTime());
    }
    @Test public void employeeRegisterRecalculatesTotalsAveragesAndEmptyState() {
        EmployeeRegister register=new EmployeeRegister("Research");
        assertTrue(register.isRegisterEmpty());assertEquals(0,register.getAverageSalary(),0.0001);
        Employee a=new Employee(new Name("Ada","Lovelace"),new Date(1,1,2020),12000);
        Employee b=new Employee(new Name("Grace","Hopper"),new Date(2,2,2021),21001);
        register.addEmployee(a);register.addEmployee(b);
        assertEquals(2,register.size());assertEquals(33001,register.getTotalSalary(),0.0001);assertEquals(16500.5,register.getAverageSalary(),0.0001);
        a.setSalary(13000);assertEquals(34001,register.getTotalSalary(),0.0001);
        assertSame(a,register.removeEmployee(0));assertSame(b,register.getEmployee(0));assertEquals(21001,register.getAverageSalary(),0.0001);
        register.clearRegister();assertTrue(register.isRegisterEmpty());assertEquals(0,register.getTotalSalary(),0.0001);assertEquals(0,register.getAverageSalary(),0.0001);
    }
}
