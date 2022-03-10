# Ingestion service to process legacy order data into Redis

# Overview

This service efficiently copies historical order data from files into REDIS, where it can be consumed by the load generator.
It is needed because consuming and modifying (masking/santizing) data from a file storage medium is slow and requires some degree of parallisation.

