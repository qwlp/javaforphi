package course.tests;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertSame;
import static org.junit.Assert.assertTrue;

import org.junit.Test;

import lib.Name;
import lib.PairOfDice;
import lib.Player;

public class PlayerFoundationTest {
	@Test
	public void defaultConstructorCreatesUsefulDefaults() {
		Player player = new Player();
		assertEquals(new Name(), player.getName());
		assertTrue(player.getPairOfDice() instanceof PairOfDice);
		assertEquals("", player.getGameTag());
	}

	@Test
	public void twoArgumentConstructorKeepsPassedObjects() {
		Name name = new Name("Joe", "Bloggs");
		String tag = new String("Invincible27");
		Player player = new Player(name, tag);
		assertSame(name, player.getName());
		assertSame(tag, player.getGameTag());
		assertTrue(player.getPairOfDice() instanceof PairOfDice);
	}

	@Test
	public void threeArgumentConstructorKeepsPassedObjects() {
		Name name = new Name("Ada", "Lovelace");
		PairOfDice dice = new PairOfDice();
		String tag = new String("AnalyticalAce");
		Player player = new Player(dice, name, tag);
		assertSame(dice, player.getPairOfDice());
		assertSame(name, player.getName());
		assertSame(tag, player.getGameTag());
	}

	@Test
	public void delegatesRollingAndScoreToPairOfDice() {
		PairOfDice dice = new PairOfDice();
		Player player = new Player(dice, new Name(), "");
		player.rollDice();
		assertEquals(dice.getScore(), player.getDiceScore());
	}

	@Test
	public void settersAndStringRepresentationWork() {
		Player player = new Player();
		Name name = new Name("Grace", "Hopper");
		String tag = new String("AmazingGrace");
		player.setName(name);
		player.setGameTag(tag);
		assertSame(name, player.getName());
		assertSame(tag, player.getGameTag());
		assertTrue(player.toString().contains("Grace"));
		assertTrue(player.toString().contains("AmazingGrace"));
	}
}
