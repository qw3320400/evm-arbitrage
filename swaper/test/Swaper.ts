import { expect } from "chai";
import { ethers } from "hardhat";
import { Web3 } from "web3";

var swaperAddress, wethAddress, aAddress, pairEAAddress, pairAEAddress, multicallAddress;

describe("Swaper", function () {

    describe("Deployment", function () {

        it("deploy Tokens", async function () {
            const accounts = await ethers.getSigners();
            const Token = await ethers.getContractFactory("Token");

            const weth = await Token.deploy("WETH", "1000000000000000000000");
            await weth.waitForDeployment();
            wethAddress = await weth.getAddress();
            expect((await weth.balanceOf(accounts[0].address)).toString()).equal("1000000000000000000000");
            console.log(`---- weth ${wethAddress}`);

            const a = await Token.deploy("A", "1000000000000000000000000000");
            await a.waitForDeployment();
            aAddress = await a.getAddress();
            expect((await a.balanceOf(accounts[0].address)).toString()).equal("1000000000000000000000000000");

        });

        it("deploy Pairs", async function () {
            const accounts = await ethers.getSigners();
            const UniswapV2Pair = await ethers.getContractFactory("UniswapV2Pair");
            const UniswapV2Factory = await ethers.getContractFactory("UniswapV2Factory");
            const Token = await ethers.getContractFactory("Token");
            const weth = await Token.attach(wethAddress);
            const a = await Token.attach(aAddress);

            const factory1 = await UniswapV2Factory.deploy();
            await factory1.waitForDeployment();
            var tx = await factory1.createPair(aAddress, wethAddress);
            await tx.wait();
            pairEAAddress = await factory1.getPair(aAddress, wethAddress);
            const pairEA = await UniswapV2Pair.attach(pairEAAddress);
            await a.transfer(pairEAAddress, "6022296110373909029866881");
            await weth.transfer(pairEAAddress, "63592353458816909596");
            await pairEA.mint(accounts[0].address);
            var reserve = await pairEA.getReserves()
            if ((await pairEA.token0()) == aAddress) {
                expect(reserve[0]).equal("6022296110373909029866881");
                expect(reserve[1]).equal("63592353458816909596");
                console.log(`---- token0 ${aAddress}`);
                console.log(`---- token1 ${wethAddress}`);
            } else {
                expect(reserve[1]).equal("6022296110373909029866881");
                expect(reserve[0]).equal("63592353458816909596");
                console.log(`---- token0 ${wethAddress}`);
                console.log(`---- token1 ${aAddress}`);
            }

            const factory2 = await UniswapV2Factory.deploy();
            await factory2.waitForDeployment();
            var tx = await factory2.createPair(aAddress, wethAddress);
            await tx.wait();
            pairAEAddress = await factory2.getPair(wethAddress, aAddress);
            const pairAE = await UniswapV2Pair.attach(pairAEAddress);
            await weth.transfer(pairAEAddress, "263943380864525275");
            await a.transfer(pairAEAddress, "20107564299619290146340");
            await pairAE.mint(accounts[0].address);
            var reserve = await pairAE.getReserves()
            if ((await pairAE.token0()) == wethAddress) {
                expect(reserve[0]).equal("263943380864525275");
                expect(reserve[1]).equal("20107564299619290146340");
                console.log(`---- token0 ${wethAddress}`);
                console.log(`---- token1 ${aAddress}`);
            } else {
                expect(reserve[1]).equal("263943380864525275");
                expect(reserve[0]).equal("20107564299619290146340");
                console.log(`---- token0 ${aAddress}`);
                console.log(`---- token1 ${wethAddress}`);
            }
        });

        it("deploy Swaper", async function () {
            const accounts = await ethers.getSigners();
            const Swaper = await ethers.getContractFactory("Swaper");
            const Token = await ethers.getContractFactory("Token");
            const weth = await Token.attach(wethAddress);

            const swaper = await Swaper.deploy(accounts[1].address, wethAddress);
            await swaper.waitForDeployment();
            swaperAddress = await swaper.getAddress();

            await weth.transfer(accounts[1].address, "1000000000000000000");
            expect((await weth.balanceOf(accounts[1].address)).toString()).equal("1000000000000000000");
        });

        it("deploy Multicall", async function () {
            const accounts = await ethers.getSigners();
            const Multicall3 = await ethers.getContractFactory("Multicall3");
            const Token = await ethers.getContractFactory("Token");
            const weth = await Token.attach(wethAddress);

            const multicall = await Multicall3.deploy();
            await multicall.waitForDeployment();
            multicallAddress = await multicall.getAddress();

            expect((await weth.balanceOf(accounts[1].address)).toString()).equal("1000000000000000000");

            await multicall.aggregate([]);
            await expect(multicall.connect(accounts[1]).aggregate([])).to.be.revertedWith("Ownable: caller is not the owner");
        });

    });

    describe("Swap", function () {

        // it("swap", async function () {
        //     const accounts = await ethers.getSigners();
        //     const Swaper = await ethers.getContractFactory("Swaper");
        //     const Token = await ethers.getContractFactory("Token");
        //     const swaper = await Swaper.attach(swaperAddress.toString());
        //     const weth = await Token.attach(wethAddress);

        //     await weth.connect(accounts[1]).approve(swaperAddress, "100000000000000000000000000");

        //     const routes = [
        //         {
        //             pair: pairCRFAWETHAddress,
        //             direction: true,
        //             amountOut: "1219469329074767527936",
        //         },
        //         {
        //             pair: pairCRFAUSDAddress,
        //             direction: false,
        //             amountOut: "11349331",
        //         },
        //         {
        //             pair: pairWETHUSDAddress,
        //             direction: false,
        //             amountOut: "7172248994489342"
        //         }
        //     ];
        //     const tx = await swaper.swap("7024483748378184", routes);
        //     const receipt = await tx.wait();
        //     const gasUsed = receipt.gasUsed;
        //     console.log(`---- gas used ${gasUsed}`)
        //     console.log(`---- balance ${(await weth.balanceOf(accounts[1].address)).toString()}`)
        // });

        it("swap2", async function () {
            const accounts = await ethers.getSigners();
            const Swaper = await ethers.getContractFactory("Swaper");
            const Token = await ethers.getContractFactory("Token");
            const swaper = await Swaper.attach(swaperAddress.toString());
            const weth = await Token.attach(wethAddress);

            await weth.connect(accounts[1]).approve(swaperAddress, "100000000000000000000000000");

            const routes = [
                {
                    pair: pairEAAddress,
                    direction: true,
                    fee: 31,
                },
                {
                    pair: pairAEAddress,
                    direction: false,
                    fee: 102,
                },
                // {
                //     pair: pairWETHUSDAddress,
                //     direction: false,
                //     fee: 30,
                // }
            ];
            const tx = await swaper.swap2("46922874771987008", routes);
            const receipt = await tx.wait();
            const gasUsed = receipt.gasUsed;
            console.log(`---- gas used ${gasUsed}`)
            console.log(`---- balance ${(await weth.balanceOf(accounts[1].address)).toString()}`)
        });

        // it("multicall", async function () {
        //     const accounts = await ethers.getSigners();
        //     const Multicall3 = await ethers.getContractFactory("Multicall3");
        //     const Token = await ethers.getContractFactory("Token");
        //     const multicall = await Multicall3.attach(multicallAddress.toString());
        //     const weth = await Token.attach(wethAddress);
        //     const web3 = new Web3;

        //     await weth.connect(accounts[1]).approve(multicallAddress, "100000000000000000000000000");

        //     const calls = [
        //         {
        //             target: wethAddress,
        //             allowFailure: false,
        //             callData: web3.eth.abi.encodeFunctionCall(
        //                 {
        //                     name: 'transferFrom',
        //                     type: 'function',
        //                     inputs: [
        //                         {
        //                             type: 'address',
        //                             name: 'from'
        //                         },
        //                         {
        //                             type: 'address',
        //                             name: 'to'
        //                         },
        //                         {
        //                             type: 'uint256',
        //                             name: 'amount'
        //                         }
        //                     ]
        //                 },
        //                 [accounts[1].address, pairCRFAWETHAddress, "7024483748378184"]
        //             )
        //         },
        //         {
        //             target: pairCRFAWETHAddress,
        //             allowFailure: false,
        //             callData: web3.eth.abi.encodeFunctionCall(
        //                 {
        //                     name: 'swap',
        //                     type: 'function',
        //                     inputs: [
        //                         {
        //                             type: 'uint256',
        //                             name: 'amount0Out'
        //                         },
        //                         {
        //                             type: 'uint256',
        //                             name: 'amount1Out'
        //                         },
        //                         {
        //                             type: 'address',
        //                             name: 'to'
        //                         },
        //                         {
        //                             type: 'bytes',
        //                             name: 'data'
        //                         }
        //                     ]
        //                 },
        //                 ["0", "1219469329074767527936", pairCRFAUSDAddress, []]
        //             )
        //         },
        //         {
        //             target: pairCRFAUSDAddress,
        //             allowFailure: false,
        //             callData: web3.eth.abi.encodeFunctionCall(
        //                 {
        //                     name: 'swap',
        //                     type: 'function',
        //                     inputs: [
        //                         {
        //                             type: 'uint256',
        //                             name: 'amount0Out'
        //                         },
        //                         {
        //                             type: 'uint256',
        //                             name: 'amount1Out'
        //                         },
        //                         {
        //                             type: 'address',
        //                             name: 'to'
        //                         },
        //                         {
        //                             type: 'bytes',
        //                             name: 'data'
        //                         }
        //                     ]
        //                 },
        //                 ["11349331", "0", pairWETHUSDAddress, []]
        //             )
        //         },
        //         {
        //             target: pairWETHUSDAddress,
        //             allowFailure: false,
        //             callData: web3.eth.abi.encodeFunctionCall(
        //                 {
        //                     name: 'swap',
        //                     type: 'function',
        //                     inputs: [
        //                         {
        //                             type: 'uint256',
        //                             name: 'amount0Out'
        //                         },
        //                         {
        //                             type: 'uint256',
        //                             name: 'amount1Out'
        //                         },
        //                         {
        //                             type: 'address',
        //                             name: 'to'
        //                         },
        //                         {
        //                             type: 'bytes',
        //                             name: 'data'
        //                         }
        //                     ]
        //                 },
        //                 ["7172248994489342", "0", accounts[1].address, []]
        //             )
        //         }
        //     ];
        //     const tx = await multicall.aggregate(calls);
        //     const receipt = await tx.wait();
        //     const gasUsed = receipt.gasUsed;
        //     console.log(`---- gas used ${gasUsed}`)
        //     console.log(`---- balance ${(await weth.balanceOf(accounts[1].address)).toString()}`)
        // });

    });

    describe("Other", function () {

        it("account", async function () {
    
        });

    });

});