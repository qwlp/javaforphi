# Aggregation

Aggregation models a “has-many” relationship. A domain class owns a collection and exposes only the operations its clients need.

## Tasks

1. Study `PlayList` as an aggregate of `Song` objects and trace its delegated add/remove/get/size operations.
2. Study `EmployeeRegister` as an aggregate of `Employee` objects.
3. Run `PlayListApp` and `EmployeeRegisterDemo`; follow the total and average calculations.
4. Compare the aggregate's focused public API with exposing an `ArrayList` directly.

Run `phi check` to compile the project.
