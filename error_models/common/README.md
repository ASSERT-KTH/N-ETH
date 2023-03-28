Note:

Because of the implementation of the fault-injection tool
the 30 error models are reduced to 19, as it is not able
to inject different types of errors for the same system call.
If there is more than 1 error in the JSON file, the tool
will pick at random, breaking our increasing aggressiveness 
assumptions.

The 19 fault injection models used for the experiments where:
1,2,3,4,5,7,9,10,11,12,13,14,15,16,17,18,19,21,25,26.
