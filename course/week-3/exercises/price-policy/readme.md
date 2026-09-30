# Price Policy Strategy

Use a common `PricePolicy` interface to change how an order line calculates its price without changing the order-line class.

## Tasks

1. Study `DefaultPricePolicy`, `DiscountPricePolicy`, and `B1G1PricePolicy`.
2. Trace delegation from `OrderLine` to its active policy.
3. Use `PricePolicyDemo` to compare results for the same product and quantity.
4. Add a policy of your own and demonstrate that `OrderLine` needs no new conditional logic.

Run `phi check` to compile the project.
