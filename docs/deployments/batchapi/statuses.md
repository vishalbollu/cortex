# Job statuses

_WARNING: you are on the master branch, please refer to the docs on the branch that matches your `cortex version`_

| Status                   | Meaning |
| :--- | :--- |
| enqueuing                | Job is being broken into batches and placed into a queue |
| running                  | Workers are retrieving batches from the queue and running inference |
| failed while enqueuing   | Failure occurred while enqueuing, check job logs for more details |
| completed with failures  | Workers completed all items in the queue but some of the batches weren't processed successfully and raised exceptions, check job logs for more details. |
| succeeded                | Workers completed all items in the queue without any failures |
| worker error             | One or more workers from an irrecoverable error, causing the job to fail, check job logs for more details. |
| out of memory            | One or more workers ran out of memory, causing the job to fail, check job logs for more details. |
| stopped                  | Job was stopped either explicitly by the user or the Batch API was deleted |