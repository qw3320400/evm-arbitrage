import { ethers } from "hardhat";

async function main() {
    // const Swaper = await ethers.getContractFactory("Swaper");
    // const swaper = await Swaper.deploy();

    // await swaper.waitForDeployment();

    // console.log(`Swaper deployed to ${swaper.getAddress()}`);
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
});
