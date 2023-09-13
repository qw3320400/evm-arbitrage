// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

interface IUniswapV2Pair {
    event Approval(address indexed owner, address indexed spender, uint value);
    event Transfer(address indexed from, address indexed to, uint value);

    function name() external pure returns (string memory);
    function symbol() external pure returns (string memory);
    function decimals() external pure returns (uint8);
    function totalSupply() external view returns (uint);
    function balanceOf(address owner) external view returns (uint);
    function allowance(address owner, address spender) external view returns (uint);

    function approve(address spender, uint value) external returns (bool);
    function transfer(address to, uint value) external returns (bool);
    function transferFrom(address from, address to, uint value) external returns (bool);

    function DOMAIN_SEPARATOR() external view returns (bytes32);
    function PERMIT_TYPEHASH() external pure returns (bytes32);
    function nonces(address owner) external view returns (uint);

    function permit(address owner, address spender, uint value, uint deadline, uint8 v, bytes32 r, bytes32 s) external;

    event Mint(address indexed sender, uint amount0, uint amount1);
    event Burn(address indexed sender, uint amount0, uint amount1, address indexed to);
    event Swap(
        address indexed sender,
        uint amount0In,
        uint amount1In,
        uint amount0Out,
        uint amount1Out,
        address indexed to
    );
    event Sync(uint112 reserve0, uint112 reserve1);

    function MINIMUM_LIQUIDITY() external pure returns (uint);
    function factory() external view returns (address);
    function token0() external view returns (address);
    function token1() external view returns (address);
    function getReserves() external view returns (uint112 reserve0, uint112 reserve1, uint32 blockTimestampLast);
    function price0CumulativeLast() external view returns (uint);
    function price1CumulativeLast() external view returns (uint);
    function kLast() external view returns (uint);

    function mint(address to) external returns (uint liquidity);
    function burn(address to) external returns (uint amount0, uint amount1);
    function swap(uint amount0Out, uint amount1Out, address to, bytes calldata data) external;
    function skim(address to) external;
    function sync() external;

    function initialize(address, address) external;
}

contract Swaper {
    // address constant account = 0x75De99Aeed9f2C11ca99906bEe209Fa03979518C;
    // address constant weth = 0x4200000000000000000000000000000000000006;

    address immutable account;
    address immutable weth;

    struct Route {
        address pair;
        bool direction;
        uint256 amountOut;
    }

    constructor (address _account, address _weth) {
        account = _account;
        weth = _weth;
    }

    function swap(uint256 amountIn, Route[] calldata routes) external {
        uint256 balance = IERC20(weth).balanceOf(account);
        if (amountIn > balance) {
            amountIn = balance;
        }
        IERC20(weth).transferFrom(account, routes[0].pair, amountIn);
        uint routesLength = routes.length;
        for (uint i = 0; i < routesLength; i++) {
            Route calldata route = routes[i];
            address to = account;
            if (i != routesLength - 1) {
                to = routes[i+1].pair;
            }
            if (route.direction) {
                IUniswapV2Pair(route.pair).swap(0, route.amountOut, to, "");
            } else {
                IUniswapV2Pair(route.pair).swap(route.amountOut, 0, to, "");
            }
        }
    }
}