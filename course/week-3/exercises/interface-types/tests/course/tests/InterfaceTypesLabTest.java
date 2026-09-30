package course.tests;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertTrue;

import java.lang.reflect.Method;

import org.junit.Test;

import lib.iterable_comparable.PlayList;
import lib.iterable_comparable.Song;
import lib.measurable.DataAnalysis;
import lib.measurable.Measurable;
import lib.polymorphism.MultipleDice;

public class InterfaceTypesLabTest {
	@Test
	public void playlistAndMultipleDiceAreIterable() {
		assertTrue("PlayList should implement Iterable", new PlayList() instanceof Iterable<?>);
		assertTrue("MultipleDice should implement Iterable", new MultipleDice() instanceof Iterable<?>);
	}

	@Test
	public void songsHaveANaturalOrderAndPlaylistCanSort() throws Exception {
		assertTrue("Song should implement Comparable", Comparable.class.isAssignableFrom(Song.class));
		Method sort = PlayList.class.getMethod("sortPlaylist");
		assertEquals(void.class, sort.getReturnType());
	}

	@Test
	public void measurableTypesExposeTheExpectedMeasurement() throws Exception {
		assertTrue(Measurable.class.isInterface());
		assertEquals(int.class, Measurable.class.getMethod("getMeasure").getReturnType());

		lib.measurable.Song song = new lib.measurable.Song("Time", 247, "Pink Floyd");
		lib.measurable.Name name = new lib.measurable.Name("Ada", "Lovelace");
		lib.measurable.Die die = new lib.measurable.Die(6);
		assertTrue(song instanceof Measurable);
		assertTrue(name instanceof Measurable);
		assertTrue(die instanceof Measurable);
		assertEquals(247, song.getMeasure());
		assertEquals(name.getFullName().length(), name.getMeasure());
		assertEquals(die.getScore(), die.getMeasure());
	}

	@Test
	public void dataAnalysisCalculatesAggregateValues() {
		DataAnalysis<lib.measurable.Song> analysis = new DataAnalysis<>();
		assertEquals(-1.0, analysis.avg(), 0.0001);
		assertEquals(-1, analysis.min());
		assertEquals(-1, analysis.max());
		analysis.addMeasurable(new lib.measurable.Song("One", 100, "A"));
		analysis.addMeasurable(new lib.measurable.Song("Two", 200, "B"));
		analysis.addMeasurable(new lib.measurable.Song("Three", 450, "C"));
		assertEquals(750, analysis.sum());
		assertEquals(250.0, analysis.avg(), 0.0001);
		assertEquals(100, analysis.min());
		assertEquals(450, analysis.max());
	}
}
