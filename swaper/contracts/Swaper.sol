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
        uint256 fee;
        // uint256 amountOut;
    }

    constructor (address _account, address _weth) {
        account = _account;
        weth = _weth;
    }

    // function swap(uint256 amountIn, Route[] calldata routes) external {
    //     uint256 balance = IERC20(weth).balanceOf(account);
    //     if (amountIn > balance) {
    //         amountIn = balance;
    //     }
    //     IERC20(weth).transferFrom(account, routes[0].pair, amountIn);
    //     uint routesLength = routes.length;
    //     Route calldata route;
    //     for (uint i = 0; i < routesLength; i++) {
    //         route = routes[i];
    //         address to = account;
    //         if (i != routesLength - 1) {
    //             to = routes[i+1].pair;
    //         }
    //         if (route.direction) {
    //             IUniswapV2Pair(route.pair).swap(0, route.amountOut, to, "");
    //         } else {
    //             IUniswapV2Pair(route.pair).swap(route.amountOut, 0, to, "");
    //         }
    //     }
    // }

    function swap2(uint256 amountIn, Route[] calldata routes) external {
        uint256 balance = IERC20(weth).balanceOf(account);
        if (amountIn > balance) {
            amountIn = balance;
        }
        uint[] memory amounts = getAmountsOut(amountIn, routes);
        require(amounts[amounts.length - 1] > amountIn, "amount out less than amount in");

        IERC20(weth).transferFrom(account, routes[0].pair, amountIn);
    
        uint routesLength = routes.length;
        Route calldata route;
        for (uint i = 0; i < routesLength; i++) {
            route = routes[i];
            address to = account;
            if (i != routesLength - 1) {
                to = routes[i+1].pair;
            }
            (uint reserveIn, uint reserveOut) = getReserves(route.pair, route.direction);
            if (route.direction) {
                IUniswapV2Pair(route.pair).swap(0, amounts[i], to, "");
            } else {
                IUniswapV2Pair(route.pair).swap(amounts[i], 0, to, "");
            }
            (reserveIn, reserveOut) = getReserves(route.pair, route.direction);
        }
        require(IERC20(weth).balanceOf(account) > balance, "balance out less than balance in");
    }

    function getAmountsOut(uint256 amountIn, Route[] calldata routes) internal view returns (uint[] memory amounts) {
        amounts = new uint[](routes.length);
        for (uint i; i < routes.length ; i++) {
            (uint reserveIn, uint reserveOut) = getReserves(routes[i].pair, routes[i].direction);
            amounts[i] = getAmountOut(amountIn, reserveIn, reserveOut, routes[i].fee);
            amountIn = amounts[i];
        }
    }

    function getReserves(address pair, bool direction) internal view returns (uint256 reserveA, uint256 reserveB) {
        (uint256 reserve0, uint256 reserve1,) = IUniswapV2Pair(pair).getReserves();
        (reserveA, reserveB) = direction ? (reserve0, reserve1) : (reserve1, reserve0);
    }

    function getAmountOut(uint256 amountIn, uint256 reserveIn, uint256 reserveOut, uint256 fee) internal pure returns (uint256 amountOut) {
        uint256 amountInWithFee = amountIn * (10000 - fee);
        uint256 numerator = amountInWithFee * reserveOut;
        uint256 denominator = reserveIn * 10000 + amountInWithFee;
        amountOut = numerator / denominator;
    }
}