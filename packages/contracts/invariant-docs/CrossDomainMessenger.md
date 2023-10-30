# `CrossDomainMessenger` Invariants

## A call to `relayMessage` should never revert if at least the proper minimum gas limits are supplied.
**Test:** [`CrossDomainMessenger.t.sol#L121`](../contracts/test/invariants/CrossDomainMessenger.t.sol#L121)

There are two minimum gas limits here:
- The outer min gas limit is for the call from the `KromaPortal` to the `L1CrossDomainMessenger`,  and it can be retrieved by calling the xdm's `baseGas` function with the `message` and inner limit.
- The inner min gas limit is for the call from the `L1CrossDomainMessenger` to the target contract.
