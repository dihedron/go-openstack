# OpenStack SDK for Golang

This project is an implementation of the OpenStack API binding for Golang.

The Gophercloud SDK has been serving the community well for quite some time but:

- lately it doesn't seem to be very active (no commits in more than three years), despite the fact that OpenStack keeps a very fast pace, releasing new features approximately every 6 months;
- looking at Gophercloud internals, there is quite a lot of clutter due to the fragmentation of the OpenStack API in the past, where some projects were embracing innovation faster than others and it was common to have multiple versions of the same shared services in use at any one time. 
 
For instance, as of today support for multiple authenticators seems like an unnecessary complication.

__This SDK is focussing on the latest version of the API.__ No backward compatibility so far.

NOTE: I have just started woking on this: it will take some time before it can be used for anything serious. At the moment only part of the OpenStack Identity V3 API is implemented.

