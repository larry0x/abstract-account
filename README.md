# abstract-account

An account abstraction solution for [CosmWasm][1]-enabled chains

## Background

There are typically two types of accounts on smart contract-enabled blockchains:

- an **externally-owned account (EOA)** is an account that is controlled by a private key. It's called "external" because the privkey isn't uploaded onchain (only the corresponding pubkey is), in contrary to–
- a **smart contract account (SCA)**, which is controlled by a binary code that is stored onchain.

Traditionally, only EOAs can initiate transactions (txs). They do so by cryptographically signing the txs using their privkeys. A tx is considered authenticated if it is accompanied by a valid signature.

In contrary, SCAs can't initiate txs because they don't have privkeys, so there's no clear way to determine whether the tx is authenticated.

Why is this bad? Because EOAs actually have very poor security properties. For one, there's no way to change the priv/pubkey associated with an EOA. This means if you lose your seed phrase, you permanently and irreversibly lose access to your account. If someone else gets hands on your seed phrase, this person permanently and irreversibly gets access to your account. In comparison, almost every web2 login allows you to change or recover your password.

Also, EOAs don't offer the flexibility to implement more sophisticated authentication logics, such as 2FA, which most web2 logins offer.

Frankly, it's hard to imagine web3 getting adoption if its account system doesn't have at least a similar level of security as web2's.

So, how do we address this? A solution put forth [by the Ethereum community][2] is **account abstraction** (AA).

Essentially, this means _to allow SCAs to initiate txs_. Authentication of txs, instead of performed at the state machine level according to a static set of rules, is hand off to the SCAs which can program whatever authentication logic it sees fit, such as allowing key rotation or 2FA. Instead of EOAs, users should choose SCAs that best suit their need.

The problem with Ethereum is that EVM is an established framework that has thousands of contracts already running on it, meaning core devs can't introduce big changes or they risk breaking many existing protocols. For Cosmos however, thanks to its [modular design][3], we're able to introduce AA without breaking anything. Let's see how this work in the next section.

## How this works

### The contract side

Remember, our goal is that instead of having the state machine performing tx authentication, we want to let the SCA do it.

In order to achieve this, the SCA must implement [two sudo methods][4], `before_tx` and `after_tx`:

```rust
enum SudoMsg {
    BeforeTx {
        msgs:       Vec<Any>,
        tx_bytes:   Binary,
        cred_bytes: Option<Binary>,
        simulate:   bool,
    },
    AfterTx {
        simulate: bool,
    },
}
```

The state machine will call `before_tx` right before a tx is about to be executed, and `after_tx` right after executing it.

- In `before_tx`, the SCA is provided with details of the tx and signing credentials. It can do signature verification here, and of course anything else it's programmed to do.

- The `after_tx` method is called only if both `before_tx` and ALL messages were executed successfully. Here it can then take any action it's programmed to. For example, if the tx includes a trade on a DEX, it can check the slippage, and reject the tx (by throwing an error) if it's above a preset threshold.

To illustrate this in a graph:

```plain
            start
              ↓
    ┌───────────────────┐
    │     before_tx     │
    └───────────────────┘
              ↓
    ┌───────────────────┐
    │        tx         │
    └───────────────────┘
              ↓
    ┌───────────────────┐
    │     after_tx      │
    └───────────────────┘
              ↓
            done
```

### The state machine side

Now let's talk about how this fit in Cosmos SDK's tx execution flow.

Each Cosmos tx consists of one or more messages (msgs). Each msg is basically an execution command, such as sending some tokens, make a delegation to a validator, call a contract, etc.

```plain
      ┌───── tx ──────┐
      │  ┌─────────┐  │
      │  │  msg 0  │  │
      │  └─────────┘  │
      │  ┌─────────┐  │
      │  │  msg 1  │  │
      │  └─────────┘  │
      │  ┌─────────┐  │
      │  │  msg 2  │  │
      │  └─────────┘  │
      └───────────────┘
```

When a tx is delivered to the state machine, it is executed by the following workflow:

- First, it runs the **AnteHandler**, which consists of a series functions to be run prior to each tx. Each such function is called a **decorator**. The Cosmos SDK provides [a number of decorators out of the box][5], including ones that set up the gas meter, verify signatures, increment the account sequence number, etc.

- Then it executes the msgs one by one.

- Finally, it runs the **PostHandler**. Similar to Ante, the PostHandler has its own set of decorators.

```plain
            start
              ↓
  ┌───── Antehandler ─────┐
  │   ┌───────────────┐   │
  │   │  decorator 0  │   │
  │   └───────────────┘   │
  │   ┌───────────────┐   │
  │   │  decorator 1  │   │
  │   └───────────────┘   │
  │   ┌───────────────┐   │
  │   │  decorator 2  │   │
  │   └───────────────┘   │
  └───────────────────────┘
              ↓
         ┌─────────┐
         │  msg 0  │
         └─────────┘
         ┌─────────┐
         │  msg 1  │
         └─────────┘
         ┌─────────┐
         │  msg 2  │
         └─────────┘
              ↓
  ┌───── Posthandler ─────┐
  │   ┌───────────────┐   │
  │   │  decorator 0  │   │
  │   └───────────────┘   │
  │   ┌───────────────┐   │
  │   │  decorator 1  │   │
  │   └───────────────┘   │
  └───────────────────────┘
              ↓
            done
```

To deploy account abstraction, we replace the default [`SigVerificationDecorator`][6], which does the authentication, with our custom [`BeforeTxDecorator`][7].

The logic of `BeforeTxDecorator` is very simple:

```python
# pseudocode
if signer_is_smart_contract(tx):
  call_smart_contract_before_tx()
else:
  run_default_sig_verification_decorator()
```

The decorator firstly determines whether the tx's signer is an SCA. If this is the case, it calls the SCA's `before_tx` method; otherwise, it simply runs the default `SigVerificationDecorator` logic.

For PostHandler, we append a new [`AfterTxDecorator`][8], where the SCA's `after_tx` method is called.

That's it - AA isn't complicated, and we don't break any existing thing. To sum it up: *two new sudo methods on the contract side, two new Ante/PostHandler decorators on the state machine side*.

## Parameters

Parameters are updatable by the module's authority, typically set to the gov module account.

- `max_gas_before` and `max_gas_after`

  Some chains may also want to limit how much gas can be consumed by the Before/AfterTx hooks. The main consideration is DoS attacks - if a malicious account has an infinite loop in on of these hooks, large amount of CPU power will be used, but since the tx fails (it uses more gas than the block gas limit) the attacker doesn't need to pay any gas fee. To prevent this, set gas limits by configuring these two params.

## How to use

See the [simapp](./simapp/) for an example.

## Demo

This repository contains three SCAs for demo purpose. Note, they are not considered ready for production use:

| Contract                                               | Description                                                  | Video         |
| ------------------------------------------------------ | ------------------------------------------------------------ | ------------- |
| [`account-asym-totp`](./cosmwasm/contracts/asym-totp/) | account that uses an asymmetric time-based one-time password | [YouTube][9]  |
| [`account-base`](./cosmwasm/contracts/base/)           | account controlled by a single secp256k1 pubkey              | n/a           |
| [`account-eth`](./cosmwasm//contracts//eth/)           | account controlled by Ethereum signatures                    | [YouTube][10] |
| [`account-granter`](./cosmwasm/contracts/granter/)     | account with authz grant capability                          | [YouTube][11] |
| [`account-updatable`](./cosmwasm/contracts/updatable/) | account with rotatable pubkey                                | [YouTube][12] |

## License

[Apache-2.0](./LICENSE)

[1]: https://cosmwasm.com/
[2]: https://eips.ethereum.org/EIPS/eip-2938
[3]: https://docs.cosmos.network/v0.46/building-modules/intro.html
[4]: https://github.com/larry0x/abstract-account/blob/main/cosmwasm/packages/abstract-account/src/lib.rs#L13-L32
[5]: https://github.com/cosmos/cosmos-sdk/blob/v0.47.2/x/auth/ante/ante.go#L38-L51
[6]: https://github.com/cosmos/cosmos-sdk/blob/v0.47.2/x/auth/ante/sigverify.go#L202-L205
[7]: https://github.com/larry0x/abstract-account/blob/main/x/abstractaccount/ante.go#L46-L128
[8]: https://github.com/larry0x/abstract-account/blob/main/x/abstractaccount/ante.go#L132-L173
[9]: https://youtu.be/XhszRNCVrpg
[10]: https://youtu.be/vI2baN2jTKY
[11]: https://youtu.be/ofB53JgsWg0
[12]: https://youtu.be/AdaLn28qG70
