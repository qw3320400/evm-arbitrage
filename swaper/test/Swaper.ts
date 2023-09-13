import { expect } from "chai";
import { ethers } from "hardhat";

var swaperAddress, wethAddress, crfaAddress, axlUSDAddress, pairCRFAWETHAddress, pairCRFAUSDAddress, pairWETHUSDAddress;

describe("Swaper", function () {

    describe("Deployment", function () {

        it("deploy Tokens", async function () {
            const accounts = await ethers.getSigners();
            const Token = await ethers.getContractFactory("Token");

            const weth = await Token.deploy("WETH", "1000000000000000000000");
            await weth.waitForDeployment();
            wethAddress = await weth.getAddress();
            expect((await weth.balanceOf(accounts[0].address)).toString()).equal("1000000000000000000000");

            const crfa = await Token.deploy("CRFA", "1000000000000000000000000");
            await crfa.waitForDeployment();
            crfaAddress = await crfa.getAddress();
            expect((await crfa.balanceOf(accounts[0].address)).toString()).equal("1000000000000000000000000");

            const axlUSD = await Token.deploy("axlUSD", "1000000000000000000000000");
            await axlUSD.waitForDeployment();
            axlUSDAddress = await axlUSD.getAddress();
            expect((await axlUSD.balanceOf(accounts[0].address)).toString()).equal("1000000000000000000000000");
        });

        it("deploy Pairs", async function () {
            const accounts = await ethers.getSigners();
            const UniswapV2Pair = await ethers.getContractFactory("UniswapV2Pair");
            const UniswapV2Factory = await ethers.getContractFactory("UniswapV2Factory");
            const Token = await ethers.getContractFactory("Token");
            const weth = await Token.attach(wethAddress);
            const crfa = await Token.attach(crfaAddress);
            const axlUSD = await Token.attach(axlUSDAddress);

            const factory = await UniswapV2Factory.deploy();
            await factory.waitForDeployment();

            var tx = await factory.createPair(crfaAddress, wethAddress);
            await tx.wait();
            pairCRFAWETHAddress = await factory.getPair(crfaAddress, wethAddress);
            const pairCRFAWETH = await UniswapV2Pair.attach(pairCRFAWETHAddress);
            await crfa.transfer(pairCRFAWETHAddress, "884916887826466518622968");
            await weth.transfer(pairCRFAWETHAddress, "5075022031094541599");
            await pairCRFAWETH.mint(accounts[0].address);
            var reserve = await pairCRFAWETH.getReserves()
            if ((await pairCRFAWETH.token0()) == crfaAddress) {
                expect(reserve[0]).equal("884916887826466518622968");
                expect(reserve[1]).equal("5075022031094541599");
            } else {
                expect(reserve[1]).equal("884916887826466518622968");
                expect(reserve[0]).equal("5075022031094541599");
            }

            var tx = await factory.createPair(crfaAddress, axlUSDAddress);
            await tx.wait();
            pairCRFAUSDAddress = await factory.getPair(crfaAddress, axlUSDAddress);
            const pairCRFAUSD = await UniswapV2Pair.attach(pairCRFAUSDAddress);
            await crfa.transfer(pairCRFAUSDAddress, "36813941190031336183629");
            await axlUSD.transfer(pairCRFAUSDAddress, "355002929");
            await pairCRFAUSD.mint(accounts[0].address);
            var reserve = await pairCRFAUSD.getReserves()
            if ((await pairCRFAUSD.token0()) == crfaAddress) {
                expect(reserve[0]).equal("36813941190031336183629");
                expect(reserve[1]).equal("355002929");
            } else {
                expect(reserve[1]).equal("36813941190031336183629");
                expect(reserve[0]).equal("355002929");
            }

            var tx = await factory.createPair(wethAddress, axlUSDAddress);
            await tx.wait();
            pairWETHUSDAddress = await factory.getPair(wethAddress, axlUSDAddress);
            const pairWETHUSD = await UniswapV2Pair.attach(pairWETHUSDAddress);
            await weth.transfer(pairWETHUSDAddress, "785015812149015823715");
            await axlUSD.transfer(pairWETHUSDAddress, "1238454830614");
            await pairWETHUSD.mint(accounts[0].address);
            var reserve = await pairWETHUSD.getReserves()
            if ((await pairWETHUSD.token0()) == wethAddress) {
                expect(reserve[0]).equal("785015812149015823715");
                expect(reserve[1]).equal("1238454830614");
            } else {
                expect(reserve[1]).equal("785015812149015823715");
                expect(reserve[0]).equal("1238454830614");
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

            await weth.transfer(accounts[1].address, "1000000000000000000")
            expect((await weth.balanceOf(accounts[1].address)).toString()).equal("1000000000000000000");
        });
    });

    describe("Swap", function () {

        it("swap", async function () {
            const accounts = await ethers.getSigners();
            const Swaper = await ethers.getContractFactory("Swaper");
            const Token = await ethers.getContractFactory("Token");
            const swaper = await Swaper.attach(swaperAddress.toString());
            const weth = await Token.attach(wethAddress);

            await weth.connect(accounts[1]).approve(swaperAddress, "100000000000000000000000000");

            const routes = [
                {
                    pair: pairCRFAWETHAddress,
                    direction: true,
                    amountOut: "1219469329074767527936",
                },
                {
                    pair: pairCRFAUSDAddress,
                    direction: false,
                    amountOut: "11349331",
                },
                {
                    pair: pairWETHUSDAddress,
                    direction: false,
                    amountOut: "7172248994489342"
                }
            ];
            const tx = await swaper.swap("7024483748378184", routes);
            const receipt = await tx.wait();
            const gasUsed = receipt.gasUsed;
            console.log(`gas used ${gasUsed}`)
            console.log(`balance ${(await weth.balanceOf(accounts[1].address)).toString()}`)
        });

    });

});