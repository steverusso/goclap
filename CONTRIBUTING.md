# Contributing Guidelines

Please create an issue before starting work on a new feature or a substantial refactor so
it can be discussed first.

## GitHub Workflow

1. Fork this repository ([if this is your first contribution](https://docs.github.com/en/get-started/quickstart/fork-a-repo)) or sync your fork ([if you've contributed before](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/syncing-a-fork)).
2. Create a new branch within your fork.
3. Commit your changes on that branch.
4. Open a pull request against this repo's `master` branch (see below for pull request guidelines).

## Pull Requests

Please try to keep pull requests simple. Overly complex code changes are difficult to
properly review. A good rule of thumb is that the pull request should focus on only one of
the following categories:

* `doc`: documentation changes only (README, repo health files, source code comments, etc)
* `feat`: implements a new feature
* `tidy`: white-space changes, reformatting, repositioning, renaming, etc.
* `refactor`: change that is not a new feature, bugfix, or tidy
* `chore`: changes that don't involve source or doc files (ex: continuous integration files)
* `examples/{name}`: changes to the examples

The title of the pull request should begin with the type of change it contains followed by
a brief overview description of what this change would accomplish. For example:
```
doc: fix typo in readme
core: bump ci versions
```

By opening a pull request to this repository, you agree to allow the project owner to
license your work under the [Unlicense](./UNLICENSE). Please only submit code where you
are the original author.

## Code Style

Generally speaking, this project adheres to the guidelines outlined in [Effective
Go](https://go.dev/doc/effective_go) and [Google's Go Style
Guide](https://google.github.io/styleguide/go/guide).
