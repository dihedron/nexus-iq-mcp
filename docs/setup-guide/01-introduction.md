# Introduction

`nexus-iq-mcp` is a Model Context Protocol (MCP) server that exposes read-only tools for the Sonatype Nexus IQ REST API. An MCP client starts it as a child process and talks to it over stdio; it needs no service unit and opens no listening port.
