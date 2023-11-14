// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { KromaToken } from "./KromaToken.sol";

contract KromaTokenMinter {
    /**
     * @notice Address of the special depositor account.
     */
    address public constant DEPOSITOR_ACCOUNT = 0xDeAddEAddeADdeADdEaDdEaddeadDeADdEAD0002;
    uint256 public constant SHARE_DENOMINATOR = 100;
    uint256 public constant MINT_DENOMINATOR = 100;
    uint256 public constant MINT_MAX_INCREASE = 3;
    uint256 public constant MINT_MIN_DECREASE = 7;
    uint256 public constant MINT_INCREASE_CAP = 10;
    uint256 public constant MINT_DECREASE_CAP = 10;

    KromaToken public immutable kromaToken;

    address[] public recipients;
    mapping(address => uint256) public shareOf;
    uint256 internal prevMinted;

    constructor(
        KromaToken _kromaToken,
        address[] memory _recipients,
        uint64[] memory _shares
    ) {
        require(_recipients.length == _shares.length, "invalid length of array");

        kromaToken = _kromaToken;

        uint256 totalShares = 0;
        for (uint256 i=0; i<_recipients.length; i++) {
            address recipient = _recipients[i];
            require(recipient != address(0), "recipient address cannot be 0");
            uint256 share = _shares[i];
            require(share != 0, "share cannot be 0");
            totalShares += share;

            recipients.push(recipient);
            shareOf[recipient] = share;
        }
        require(totalShares == SHARE_DENOMINATOR, "invalid shares");
    }

    function mint(uint256 totalAmount) external {
        require(msg.sender == DEPOSITOR_ACCOUNT, "only depositor can call this function");

        // TODO(pangssu): add distribution logic
        for (uint256 i=0; i<recipients.length; i++) {
            address recipient = recipients[i];
            uint256 share = shareOf[recipient];
            uint256 amount = totalAmount * share / SHARE_DENOMINATOR;

            kromaToken.mint(recipient, amount);
        }
    }
}
