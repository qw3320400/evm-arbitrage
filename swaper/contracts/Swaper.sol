// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

interface IERC20 {
    function balanceOf(address account) external view returns (uint256);
    function transferFrom(address from, address to, uint256 amount) external returns (bool);
}

interface IUniswapV2Pair {
    function getReserves() external view returns (uint112 reserve0, uint112 reserve1, uint32 blockTimestampLast);
    function swap(uint amount0Out, uint amount1Out, address to, bytes calldata data) external;
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

    function swap(bytes memory params) external {
        uint80 amountIn;
        uint8 length;
        assembly {
            length := div(sub(mload(params), 10), 23)
            amountIn := mload(add(params, 10))
        }
        Route[] memory routes = new Route[](length);
        address pair;
        uint8 direction;
        uint16 fee;
        for (uint8 i = 0; i < length; i++) {
            assembly {
                let pointer := add(params, add(30, mul(i, 23)))
                pair := mload(pointer)
                pointer := add(pointer, 1)
                direction := mload(pointer)
                pointer := add(pointer, 2)
                fee := mload(pointer)
            }
            routes[i].pair = pair;
            routes[i].direction = direction == 0 ? false : true ;
            routes[i].fee = fee;
        }
        _swap(amountIn, routes);
    }

    function _swap(uint256 amountIn, Route[] memory routes) internal {
        uint256 balance = IERC20(weth).balanceOf(account);
        if (amountIn > balance) {
            amountIn = balance;
        }
        uint[] memory amounts = getAmountsOut(amountIn, routes);
        require(amounts[amounts.length - 1] > amountIn, "amount out less than amount in");

        IERC20(weth).transferFrom(account, routes[0].pair, amountIn);
    
        uint routesLength = routes.length;
        Route memory route;
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

    function getAmountsOut(uint256 amountIn, Route[] memory routes) internal view returns (uint[] memory amounts) {
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