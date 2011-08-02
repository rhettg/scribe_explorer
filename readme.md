Scribe Explorer
==============
This project was original developed for Yelp's 5th Hackathon by the team of rhettg and goldenberg. It's our first real golang project and done over about 36 hours, so forgive the lack of production value.

Overview
-----------
We generate several log streams in JSON format that move through our datacenter by using scribe. As part of our scribe infrastructure, there is a live tailing service that allows a process to choose a particular scribe log and stream all the data in near real time.

At Yelp, we have several processes handling this stream so as to provide real time services such as monitoring and stats generation. Often we do ad-hoc analysis of this data stream as well to watch specific log elements during a push or doing some debugging etc.

The firehouse of information is very resource intensive. Processing a single real-time stream can make a noticable impact on the production machine that has to handle it.

What Problem Does this Solve
-----------------------------
Scribe Explorer hopefully:

  * Provides a mechanism for doing multiple real-time analysis projects with much less resource overhead.
  * Makes doing ad-hoc analysis easier by providing a web interface for simple processing


Web Interface
-------------
The web interface provides:

  * Selection of log file to process
  * Fields to display (in the format A.0.foo in an object such as {'A': [{'foo': True}]}
  * Filter to apply (such as 'servlet == home' and 'sample 0.25'

The web interface will then stream the resulting data and display the most recent page of data in tabular form. The stream may be stopped by hitting the 'Stop' button. Or a new query can be started at any time.

The interface is found on localhost:8080

Raw Interface
-------------

Similiar to the web interface, scripts may use Scribe Explorer to programmatically setup a custom processed stream. An example script is found in utils/tailer.py

The interface is found on localhost:3535

Building And Installing
----------------------
There is a Makefile. Typically it should be built as:

    gomake

There is no configuration right now, and the URL for the tailing service as well as the bind host and ports are hard coded.

Future Work
-----------
  * Aggregation functions seem to have some issues. At least during our demo.
  * Displaying complex objects show up as 'Object'. Last minute attempts to display JSON formatted or pretty printed didn't quite work out.
  * After all clients end their subscriptions to a given stream, the stream should disconnect. This would ensure resource usage is near 0 when there are no subscribers.
  * Web interface should provide some way to examine the full entry of some sample log entries. We should also be able to generate a list of log streams.
  * Productionize
    * Configuration files
    * Tests
    * Monitoring actions / status
    * Ensure logging is useful



