#!/bin/bash

# Process each line, handling both JSON and non-JSON output
while IFS= read -r line; do
  # Try to parse as JSON, if it fails just print the line as-is
  echo "$line" | jq -r '
    def colorize(color; text): 
      if color == "red" then "\u001b[31m" + text + "\u001b[0m"
      elif color == "yellow" then "\u001b[33m" + text + "\u001b[0m"
      elif color == "green" then "\u001b[32m" + text + "\u001b[0m"
      elif color == "blue" then "\u001b[34m" + text + "\u001b[0m"
      else text
      end;
    
    def severity_color:
      if .severity == "ERROR" then "red"
      elif .severity == "WARN" then "yellow"
      elif .severity == "INFO" then "green"
      else "blue"
      end;
    
    "\(.time) " + 
    colorize(severity_color; "[\(.severity)]") + 
    " \(.message)" +
    (if .error then " ERROR: " + colorize("red"; .error) else "" end) +
    (if ."correlation-id" then " [ID: \(."correlation-id")]" else "" end)
  ' 2>/dev/null || echo "$line"
done