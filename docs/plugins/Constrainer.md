# Constrainer

There may be some cases where you want to prevent some combination of Chaos
Monkey terminations, but the [configuration options](../Configuring-behavior-via-sysbreaker) aren't flexible
enough for your use case.

You can define a custom constrainer to do this.

As an example, let's say you wanted to disallow any terminations for apps
that contain "foo" as a substring.

```go
package constrainer

import (
	"github.com/FakeTwitter/elon/deps"
	"github.com/FakeTwitter/elon/config"
	"github.com/FakeTwitter/elon/schedule"
    "strings"
)

func init() {
    deps.GetConstrainer = getConstrainer()
}

type noFoo struct {}

func getConstrainer(cfg *config.Monkey) (schedule.Constrainer, error) {
    return noFoo{}, nil
}

func (n noFoo) Filter(s schedule.Schedule) schedule.Schedule {
	result := schedule.New()
	for _, entry := range s.Entries() {
        if !strings.Contains(entry.Group.Team(), "foo") {
            result.Add(entry.Time, entry.Group)
        }
    }
    return result
}

```

See the [Plugins](index.md) page for info on how to build a custom version of
Elon with your plugin.
