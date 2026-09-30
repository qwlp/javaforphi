package phi.tests;
import static org.junit.Assert.*;
import org.junit.Test;
import lib.*;

public class CountableTest {
    @Test public void counterSupportsPositiveNegativeAndZeroCountsThroughInterface() {
        for(int initial:new int[]{-7,0,1,17}) {
            Counter counter=new Counter(initial);Countable value=counter;
            assertEquals(initial,value.getCount());value.increment();assertEquals(initial+1,value.getCount());value.decrement();value.decrement();assertEquals(initial-1,value.getCount());counter.reset();assertEquals(0,value.getCount());
        }
    }
    @Test public void moduloCounterWrapsBothDirectionsAtEveryBoundary() {
        for(int modulo:new int[]{1,2,7,10,13}) {
            ModuloCounter counter=new ModuloCounter(modulo);Countable value=counter;
            value.decrement();assertEquals("Zero must wrap to modulo minus one",modulo-1,value.getCount());value.increment();assertEquals(0,value.getCount());
            for(int i=0;i<modulo*3;i++) {value.increment();assertEquals((i+1)%modulo,value.getCount());}
            counter.reset();assertEquals(0,value.getCount());
        }
    }
    @Test public void changingModuloNormalizesCurrentCount() {
        ModuloCounter counter=new ModuloCounter(19,13);assertEquals(6,counter.getCount());counter.setModulo(4);assertEquals(2,counter.getCount());assertEquals(4,counter.getModulo());
    }
}
