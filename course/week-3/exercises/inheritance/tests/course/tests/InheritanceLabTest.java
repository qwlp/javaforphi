package course.tests;

import static org.junit.Assert.assertEquals;
import static org.junit.Assert.assertFalse;
import static org.junit.Assert.assertNotEquals;
import static org.junit.Assert.assertTrue;

import org.junit.Test;

import lib.bankaccounts.BankAccount;
import lib.bankaccounts.InterestAccount;
import lib.bankaccounts.IsaAccount;
import lib.bankaccounts.StudentAccount;
import lib.counters.Counter;
import lib.counters.StepCounter;

public class InheritanceLabTest {
	@Test
	public void stepCounterIsACounterAndUsesItsStep() {
		StepCounter counter = new StepCounter();
		assertTrue(counter instanceof Counter);
		counter.setStep(3);
		assertEquals(3, counter.getStep());
		counter.increment();
		assertEquals(3, counter.getCount());
		counter.decrement();
		assertEquals(0, counter.getCount());
		assertTrue(counter.toString().contains("step=3"));
	}

	@Test
	public void studentAccountHonorsItsOverdraftLimit() {
		StudentAccount account = new StudentAccount();
		assertTrue(account instanceof BankAccount);
		account.setOverdraftLimit(1000);
		assertEquals(1000, account.getOverdraftLimit());
		account.withdraw(1000);
		assertEquals(-1000, account.getBalance());
		account.withdraw(1);
		assertEquals(-1000, account.getBalance());
		assertTrue(account.toString().contains("overdraftLimit=1000"));
	}

	@Test
	public void isaAccountHonorsAndResetsItsDepositAllowance() {
		IsaAccount account = new IsaAccount();
		assertTrue(account instanceof InterestAccount);
		account.resetDepositRemaining();
		int allowance = account.getDepositRemaining();
		assertTrue("reset allowance should be positive", allowance > 0);
		account.deposit(allowance);
		assertEquals(0, account.getDepositRemaining());
		int balance = account.getBalance();
		account.deposit(1);
		assertEquals(balance, account.getBalance());
		assertTrue(account.toString().contains("depositRemaining=0"));
	}

	@Test
	public void accountEqualityIncludesExactClassAndSubclassState() {
		assertEquals(new BankAccount(100), new BankAccount(100));
		assertNotEquals(new BankAccount(100), new BankAccount(101));
		assertFalse(new BankAccount(100).equals(new InterestAccount(100, 0, 0)));

		StudentAccount first = new StudentAccount();
		StudentAccount second = new StudentAccount();
		first.setOverdraftLimit(50);
		second.setOverdraftLimit(50);
		assertEquals(first, second);
		second.setOverdraftLimit(51);
		assertNotEquals(first, second);
	}
}
