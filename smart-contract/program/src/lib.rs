use borsh::{BorshDeserialize, BorshSerialize};
use solana_program::{
    account_info::{next_account_info, AccountInfo},
    entrypoint,
    program_error::ProgramError,
    pubkey::Pubkey,
};

#[derive(BorshSerialize, BorshDeserialize, Debug)]
pub struct ZkpResult {
    pub proof_json: Vec<u8>,
    pub verifying_key: Vec<u8>,
    pub public_witness: Vec<u8>,
}

#[cfg(not(feature = "exclude_entrypoint"))]
entrypoint!(process_instruction);

pub fn process_instruction(
    program_id: &Pubkey,
    accounts: &[AccountInfo],
    instruction_data: &[u8],
) -> entrypoint::ProgramResult {
    let accounts_iter = &mut accounts.iter();
    let account = next_account_info(accounts_iter)?;

    if account.owner != program_id {
        return Err(ProgramError::IncorrectProgramId);
    }

    let zkp_result = ZkpResult::try_from_slice(instruction_data)?;
    let mut account_data = account.data.borrow_mut();
    zkp_result.serialize(&mut *account_data)?;
    Ok(())
}
