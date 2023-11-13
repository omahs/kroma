// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { ERC20Burnable } from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import { ERC20Capped } from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Capped.sol";
import { ERC20Permit } from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";
import { ERC20Votes } from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Votes.sol";

contract KromaToken is ERC20Burnable, ERC20Capped, ERC20Votes, Ownable {
    /**
     * @notice Constructs the Kroma token contract.
     */
    constructor() ERC20("Kroma", "KRO") ERC20Capped(50_000_000 ether) ERC20Permit("Kroma") {}

    /**
     * @notice Allows the owner to mint tokens.
     *
     * @param _account The account receiving minted tokens.
     * @param _amount  The amount of tokens to mint.
     */
    function mint(address _account, uint256 _amount) public onlyOwner {
        _mint(_account, _amount);
    }

    /**
     * @notice Callback called after a token transfer.
     *
     * @param from The account sending tokens.
     * @param to     The account receiving tokens.
     * @param amount The amount of tokens being transferred.
     */
    function _afterTokenTransfer(
        address from,
        address to,
        uint256 amount
    ) internal override(ERC20, ERC20Votes) {
        super._afterTokenTransfer(from, to, amount);
    }

    /**
     * @notice Internal mint function.
     *
     * @param to     The account receiving minted tokens.
     * @param amount The amount of tokens to mint.
     */
    function _mint(address to, uint256 amount) internal override(ERC20, ERC20Capped, ERC20Votes) {
        super._mint(to, amount);
    }

    /**
     * @notice Internal burn function.
     *
     * @param account The account that tokens will be burned from.
     * @param amount  The amount of tokens that will be burned.
     */
    function _burn(address account, uint256 amount) internal override(ERC20, ERC20Votes) {
        super._burn(account, amount);
    }
}
