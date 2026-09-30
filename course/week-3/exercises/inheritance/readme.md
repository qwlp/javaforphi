# Inheritance

Reuse and specialize behavior through `extends`, overridden methods, `super`, and polymorphism.

## Tasks

1. Create `StepCounter extends Counter`, with a step field, constructors, accessors, stepped increment/decrement, and `toString`.
2. Create `StudentAccount extends BankAccount`; withdrawals may reach but never exceed its overdraft limit.
3. Create `IsaAccount extends InterestAccount`; deposits reduce a remaining allowance and are rejected if they exceed it. Add getter, reset-to-a-positive-fixed-amount, and `toString`.
4. Implement exact-class, state-based `equals` in every bank-account level as described in the lab.

Run `phi check`. Tests cover specialization, boundary behavior, representation, and equality.
