// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { KromaToken } from "./KromaToken.sol";

contract KromaTokenMinter {
    /**
     * @notice Address of the special depositor account.
     */
    address public constant DEPOSITOR_ACCOUNT = 0xDeAddEAddeADdeADdEaDdEaddeadDeADdEAD0002;

    uint256 public constant MINT_DENOMINATOR = 100;
    uint256 public constant MINT_MAX_INCREASE = 3;
    uint256 public constant MINT_MIN_DECREASE = 7;
    uint256 public constant MINT_INCREASE_CAP = 10;
    uint256 public constant MINT_DECREASE_CAP = 10;

    KromaToken public immutable kromaToken;
    address public immutable investorVault;
    address public immutable teamVault;
    address public immutable wemixFoundationVault;
    address public immutable wemixCommunityVault;
    address public immutable airdropVault;
    address public immutable partnerVault;
    address public immutable utilityVault;
    address public immutable securityCouncilVault;
    address public immutable communityImpactVault;
    address public immutable validatorVault;

    uint256 internal prevMinted;

    error ErrNoDepositor();

    modifier onlyDepositor() {
        if (msg.sender != DEPOSITOR_ACCOUNT) {
            revert ErrNoDepositor();
        }
        _;
    }

    constructor(
        KromaToken _kromaToken,
        address _investorVault,
        address _teamVault,
        address _wemixFoundationVault,
        address _wemixCommunity,
        address _airdropVault,
        address _partnerVault,
        address _utilityVault,
        address _securityCouncilVault,
        address _communityImpactVault,
        address _validatorReward
    ) {
        kromaToken = _kromaToken;
        investorVault = _investorVault;
        teamVault = _teamVault;
        wemixFoundationVault = _wemixFoundationVault;
        wemixCommunityVault = _wemixCommunity;
        airdropVault = _airdropVault;
        partnerVault = _partnerVault;
        utilityVault = _utilityVault;
        securityCouncilVault = _securityCouncilVault;
        communityImpactVault = _communityImpactVault;
        validatorVault = _validatorReward;
    }

    function mint() external onlyDepositor {
        // TODO(pangssu): add distribution logic
        kromaToken.mint(investorVault, 1 ether);
        kromaToken.mint(teamVault, 1 ether);
        kromaToken.mint(wemixFoundationVault, 1 ether);
        kromaToken.mint(wemixCommunityVault, 1 ether);
        kromaToken.mint(partnerVault, 1 ether);
        kromaToken.mint(utilityVault, 1 ether);
        kromaToken.mint(securityCouncilVault, 1 ether);
        kromaToken.mint(communityImpactVault, 1 ether);
        kromaToken.mint(validatorVault, 1 ether);
    }
}
