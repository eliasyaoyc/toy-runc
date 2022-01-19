# Go-workQueue

**WorkQueue** is an easy-to-use utility that provides a simple queue that supports the following features:
* Element processed in the order in which they are added.
* A single element will not be processed multiple times concurrently, and if an element's added multiple times before it can be processed, it will only be processed once.
* Multiple consumers and producers. In particular, it is allowed for an item to be reenqueued while it is being processed.
* Shutdown notifications.


Licensed under of either of

* MIT license ([LICENSE-MIT](LICENSE) or http://opensource.org/licenses/MIT)
