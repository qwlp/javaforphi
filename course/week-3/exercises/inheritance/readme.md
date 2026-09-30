# Inheritance

## Goal

Reuse and specialize behavior through `extends`, overridden methods, `super`, and polymorphism.

## Files to work on

Starter source folders: `src/lib/bankaccounts/`, `src/lib/counters/`, `src/main/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Tasks

1. Create `StepCounter extends Counter`, with a step field, constructors, accessors, stepped increment/decrement, and `toString`.
2. Create `StudentAccount extends BankAccount`; withdrawals may reach but never exceed its overdraft limit.
3. Create `IsaAccount extends InterestAccount`; deposits reduce a remaining allowance and are rejected if they exceed it. Add getter, reset-to-a-positive-fixed-amount, and `toString`.
4. Implement exact-class, state-based `equals` in every bank-account level as described in the lab.

## Expected behavior

StepCounter uses its step; StudentAccount enforces its overdraft boundary; IsaAccount enforces its remaining allowance; account equality considers exact class and state.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The behavioral tests must pass. Use failures to identify the method or boundary case to revisit.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint: open if you get stuck</summary>

Use super to reuse parent behavior, then apply the subclass rule. Check exact withdrawal and deposit limits as well as values just beyond them.

</details>
