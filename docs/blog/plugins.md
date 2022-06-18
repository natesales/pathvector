---
slug: extending-pathvector-with-plugins

title: Extending Pathvector with Plugins

author: Nate Sales

author_title: Developer

author_url: https://github.com/natesales

author_image_url: https://github.com/natesales.png

date: June 20, 2022

tags: [v6-launch]
---

*This post is part of our [version 6 launch series](/blog/tags/v-6-launch)!*

Extending Go programs is hard.

## One-upping the standard library

The Caddy web server uses a complex but well-designed [plugin system](https://caddyserver.com/docs/extending-caddy)
complete with a custom [build tool](https://github.com/caddyserver/xcaddy) and a
dynamic [download page](https://caddyserver.com/download) that allows users to choose plugins before downloading a
binary. CoreDNS also [adopted this design](https://coredns.io/2017/03/01/how-to-add-plugins-to-coredns/) for their
extensible DNS server.

Developing a plugin with the Caddy/CoreDNS design like this:

1. Create a new Go package with the standard go.mod file and an arbitrary package name (not `main`)
2. Import the Caddy plugin interface from Caddy's own Go package
3. Create an exported structure with whatever fields you like
4. Implement plugin functions as outlined by the plugin interface
5. Call the `RegisterModule` function on your struct
6. Write your plugin logic
7. Add an interface guard to validate that your plugin struct
8. Build Caddy with your plugin

By this point your plugin is baked into the Caddy binary, and you're off to the races. Internally, Caddy uses Go's code
generation feature to modify its source code to "underscore import" your plugin package (i.e. importing the plugin
package only for side effects, without calling any functions).

This design is neat because the resulting binary contains all the plugin code and therefore doesn't need any external
object files.

The obvious downside is that you need to replace the entire program any time you want to modify a plugin. In other
words, you need to build a brand new binary even if you're only modifying a tiny part of a plugin. The second problem is
distributing the program. Caddy addresses this issue with what amounts to a public-facing build server; to download a
pluginized version of Caddy, you choose from a list of packages on their download site and a build server responds with
a binary to match. *Binaries cooked to order!*

## A little history

Pathvector used this same design (without the custom build server) when we introduced plugin support in 5.6.1 and
enabled Pathvector at the edge of Kubernetes clusters with
the [Calico plugin](https://github.com/natesales/pathvector-plugin-calico-5). In addition to the aforementioned issues,
the binary was huge due to the size of the Kubernetes Go API and without a fancy build server like Caddy, updates were
painful.

## Standard library to the rescue?

If rolling your own plugin system is so hard, what about using the language's own features? Go has native plugin support
via it's [plugin](https://pkg.go.dev/plugin) package in the standard library.

Unfortunately, with simplicity comes a slew of issues:

- The plugin and program must be compiled with identical Go versions
- The plugin and program must have identical 3rd party package versions
- The plugin and program must have an identical GOPATH
- ... and a whole bunch of libc linker intricacies that break cross compilation

## What if plugins were more like microservices?

Hashicorp has an interesting [plugin system](https://github.com/hashicorp/go-plugin) that uses RPC for communication
rather than linking or source modification.
