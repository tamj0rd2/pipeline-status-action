# pipeline-status-action

This action polls the github status API for the given statuses to complete successfully and sends a message on slack
if any of the checks have failed or do not complete within the specified timeout.

## Inputs

Take a look at [./action.yaml](./action.yaml) for the full list of inputs and defaults etc

## Example usage

```yaml
on:
  push:
    branches:
      - main

jobs:
  report-failed-commits:
    runs-on: ubuntu-latest
    steps:
      - name: Check pipeline statuses
        uses: tamj0rd2/pipeline-status-action@v1.1.0
        with:
          checkNames: statusName1,status with spaces in the name,another-status-name
```
