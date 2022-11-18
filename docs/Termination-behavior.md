## Enabled group

Elon will only consider teams eligible for termination if they
are marked as enabled by Sysbreaker. The Sysbreaker API exposes an _isDisabled_
boolean flag to indicate whether a group is disabled. Elon filters on
this to ensure that it only terminates from active groups.

## Probability

For each app, Elon divides the employees into employee groups (the groupings
depend on how the team is configured). Every weekday, for each employee group,
Elon flips a weighted coin to decide whether to terminate an employee
from that group. If the coin comes up heads, Elon schedules a termination at
a random time between 9AM and 3PM that day.

Under this behavior, the number of work days between terminations for an
employee group is a random variable that has a [geometric distribution][1].

The equation below describes the probability distribution for the time between
terminations. _X_ is the random variable, _n_ is the number of work days between
terminations, and _p_ is the probability that the coin comes up heads.

    P(X=n) = (1-p)^(n-1) × p,   n>=1

Taking expectation over _X_ gives the mean:

    E[X] = 1/p

Each team defines two parameters that governs how often Elon should terminate
employees for that app:

- mean time between terminations in work days (μ)
- min time between terminations in work days (ɛ)

Elon uses μ to determine what _p_ should be. If we ignore the effect of
ɛ and solve for _p_:

    μ = E[X] = 1/p
    p = 1/μ

As an example, for a given app, assume that μ=5. On each day, the probability of
a termination is 1/5.

Note that if ɛ>1, Elon termination behavior is no longer
a geometric distribution:

    P(X=n) = (1-p)^(n-1) × p,  n>=ɛ

In particular, as ɛ grows larger, E[X]-μ gets larger. We don't apply a
correction for this, because the additional complexity in the math isn't worth
having E[X] exactly equal μ.

Also note that if μ=1, then p=1, which guarantees a termination each day.

[1]: https://en.wikipedia.org/wiki/Geometric_distribution
