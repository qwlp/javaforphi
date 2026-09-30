# Price Policy Strategy

## Goal

Use a common `PricePolicy` interface to change how an order line calculates its price without changing the order-line class.

## Files to work on

Starter source folders: `src/lib/`, `src/main/`. Follow the tasks below for the classes to edit or create. The copied Word handout contains the full lab specification.

## Tasks

1. Study `DefaultPricePolicy`, `DiscountPricePolicy`, and `B1G1PricePolicy`.
2. Trace delegation from `OrderLine` to its active policy.
3. Use `PricePolicyDemo` to compare results for the same product and quantity.
4. Add a policy of your own and demonstrate that `OrderLine` needs no new conditional logic.

## Expected behavior

The same order line computes different prices when its policy changes. A new policy works through the existing interface without editing OrderLine.

## Automated check requirements

Checks cover zero, odd, and even quantities; fractional discounts; and policy replacement. `OrderLine.getCost()` must delegate to the stored `PricePolicy`, including a policy implementation the checker supplies. No particular console demo format is required.

## Check your work

Run `phi check` from this lesson folder (or any nested source folder). The embedded behavioral and boundary tests must pass. Compilation is only the first stage; a compiling but incorrect implementation fails.

For an optional clue, run `phi hint`.

Once you have checked the tasks, run `phi submit` to record local completion, then `phi next` to continue.

<details>
<summary>Hint: open if you get stuck</summary>

Keep pricing rules inside policy classes. Try the same quantity against each policy to see how the strategies differ.

</details>
