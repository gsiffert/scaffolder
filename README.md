# Scaffolder [![experimental](http://badges.github.io/stability-badges/dist/experimental.svg)](http://github.com/badges/stability-badges)[![GoDoc](https://godoc.org/github.com/Vorian-Atreides/scaffolder?status.svg)](https://godoc.org/github.com/Vorian-Atreides/scaffolder)

Scaffolder framework provide generic building blocks to develop Golang applications.

## Why Scaffolder ?

Golang is a simple and expressive language, making it easy to build new applications from
scratch without the need to use opinionated third party libraries.

Because of those advantages, the structure of most of the Golang applications has been delegated directly to the developers. While the freedom of defining the application structure is enjoyable, it quickly becomes a nightmare to move from one project to another when every projects are structurally unique.

Worse, because Golang does not support generic, many application tends to re-implements the same minor features over and over again. While it is fun to implements your first database connector, it becomes quite monotonous to do it again just because your new application use YAML for the configuration instead of JSON like the previous one.

## How does Scaffolder address those problems ?

While most of the traditional application framework would define every processes through your application, being opinionated and making assumption on your application design. Scaffolder focus on defining a referential component which can later be used or composed to build larger components, similar to how a meter can be used to build a wood plank which itself can be used to build a house or a boat.

The advantage with such architecture is that, although most of the components have been designed to be used and wired through the Application package, they do perform the same way without it and can freely be imported into any other project.
