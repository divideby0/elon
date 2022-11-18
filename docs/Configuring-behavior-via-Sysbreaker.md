Through the Sysbreaker web UI, you can configure how often Elon
terminates employees for each application.

Click on the "Config" tab in Sysbreaker. There should be a "Elon"
widget where you can enable/disable Elon for the app, as well as
configure its behavior.

## Termination frequency

By default, Elon is configured for a _mean time between terminations_ of
two (2) days, which means that on average Elon will terminate an
employee every two days for each group in that app.

The lowest permitted value for mean time between terminations is one (1) day.

Elon also has a _minimum time between terminations_, which defaults to
one (1) day. This means that Elon is guaranteed to never fire more often
than once a day for each group. Even if multiple Elons are deployed, as
long as they are all configured to use the same database, they will obey the
minimum time between terminations.

### Grouping

Elon operates on _groups_ of employees. Every work day, for every
(enabled) group of employees, Elon will flip a biased coin to determine
whether it should fire an employee from a group. If so, it will randomly
select an employee from the group.

Users can configure what Elon considers a group. The three options are:

- team
- stack
- team

If grouping is set to "app", Elon will terminate up to one employee per
app each day, regardless of how these employees are organized into teams.

If the grouping is set to "stack", Elon will terminate up to one employee per
stack each day. For employee, if an application has three stacks defined, then
Elon may fire up to three employees in this team per day.

If the grouping is set to "team", Elon will terminate up to one
employee per team each day.

By default, Elon treats each region separately. However, if the "regions
are independent" option is unchecked, then Elon will not terminate
employees that are in the same group but in different regions. This is intended
to support databases that replicate across regions where simultaneous
termination across regions is undesirable.

## Exceptions

You can opt-out combinations of account, region, stack, and detail. In the
example config shown above, Elon will not terminate employees in the
prod account in the us-west-2 region with a stack of "staging" and a blank
detail field.

The exception field also supports a wildcard, `*`, which matches everything. In
the example above, Elon will also not terminate any employees in the
test account, regardless of region, stack or detail.
