import { expect } from "chai";
import { ethers } from "hardhat";
import { Web3, eth } from "web3";

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

        // it("swap2", async function () {
        //     const accounts = await ethers.getSigners();
        //     const Swaper = await ethers.getContractFactory("Swaper");
        //     const Token = await ethers.getContractFactory("Token");
        //     const swaper = await Swaper.attach(swaperAddress.toString());
        //     const weth = await Token.attach(wethAddress);

        //     await weth.connect(accounts[1]).approve(swaperAddress, "100000000000000000000000000");

        //     const routes = [
        //         {
        //             pair: pairEAAddress,
        //             direction: true,
        //             fee: 31,
        //         },
        //         {
        //             pair: pairAEAddress,
        //             direction: false,
        //             fee: 102,
        //         },
        //     ];
        //     const tx = await swaper.swap2("46922874771987008", routes);
        //     const receipt = await tx.wait();
        //     const gasUsed = receipt.gasUsed;
        //     console.log(`---- gas used ${gasUsed}`) // 239673
        //     console.log(`---- balance ${(await weth.balanceOf(accounts[1].address)).toString()}`)
        // });

        it("swap", async function () {
            const accounts = await ethers.getSigners();
            const Swaper = await ethers.getContractFactory("Swaper");
            const Token = await ethers.getContractFactory("Token");
            const swaper = await Swaper.attach(swaperAddress.toString());
            const weth = await Token.attach(wethAddress);

            await weth.connect(accounts[1]).approve(swaperAddress, "100000000000000000000000000");

            var param = ethers.concat([
                ethers.zeroPadValue(ethers.toBeHex("46922874771987008"), 10),
                ethers.zeroPadValue(ethers.toBeHex(pairEAAddress), 20),
                ethers.toBeHex(1),
                ethers.zeroPadValue(ethers.toBeHex(31), 2),
                ethers.zeroPadValue(ethers.toBeHex(pairAEAddress), 20),
                ethers.toBeHex(0),
                ethers.zeroPadValue(ethers.toBeHex(102), 2),
            ])
            console.log(param);
            
            const tx = await swaper.swap(param);
            const receipt = await tx.wait();
            const gasUsed = receipt.gasUsed;
            console.log(`---- gas used ${gasUsed}`)
            console.log(`---- balance ${(await weth.balanceOf(accounts[1].address)).toString()}`)
        });

        it("getAmountOut", async function () {
            // const Swaper = await ethers.getContractFactory("Swaper");
            // const swaper = await Swaper.attach(swaperAddress.toString());

            // var param = ethers.hexlify("0x000000003ba4c0603f8c55000adfc0babd65f1c43a1eb290156931d4b67301001eddceda9866c0bced84561673ddef77d218b6d90e0000193cd5a4c56c4794d88170b6bde719656a7302653900001cc52328d5af54a12da68459ffc6d0845e91a8395f00001aa76ebd17353035d84560eec5eff22de533f0052c01006d");
            // param = ethers.concat([param]);
            // console.log(param);
            // await swaper.testGetAmountOut(param);
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