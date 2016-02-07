# Contributing

We're always open to new issues or pull requests.

## Find a bug?

Create a [new issue](https://github.com/Tonkpils/snag/issues/new) with
a brief explanation of the problem. Please include your snag file and logs
to help us best diagnose the issue.

## Have a feature request?

Snag is ever evolving and we want to make sure everyone
can get as much use out of the tool as we did. If there
is something you'd like snag to be able to do create a
[new issue](https://github.com/Tonkpils/snag/issues/new)
with a description of the feature you're looking for.

## Writing code

### Setting up your environment

In order to build and test snag properly, you need to use go1.5+
with the `GO15VENDOREXPERIMENT` environment variable set to `1`.


#### Optional

Before you start writing code, we suggest that you
build a version of snag with all the changes in the
master branch that you can use to test your changes with:

```bash
go build -o snagMaster
```

Afterward just run `./snagMaster` and develop away!

### Creating a Pull Request

We *highly* recommend using snag to build, lint and test
your code before submitting a pull request since our snag file
is very similar to our travis file. We will try to at least comment
on pull requests within a couple of days. We may suggest
some changes or improvements or alternatives.

Some things that will increase the chance that your pull request is accepted:

* Write tests.
* Format your code.
* Write a [good commit message][commit].

[commit]: http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html