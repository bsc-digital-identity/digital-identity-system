use borsh::{BorshDeserialize, BorshSerialize};
use solana_program::{
    account_info::{next_account_info, AccountInfo},
    entrypoint,
    msg,
    program_error::ProgramError,
    pubkey::Pubkey,
};

#[derive(BorshSerialize, BorshDeserialize, Debug)]
pub struct ZkpResult {
    pub proof: Vec<u8>,
    pub verifying_key: Vec<u8>,
    pub public_witness: Vec<u8>,
}

#[cfg(not(feature = "exclude_entrypoint"))]
entrypoint!(process_instruction);

pub fn process_instruction(
    _program_id: &Pubkey,
    accounts: &[AccountInfo],
    instruction_data: &[u8],
) -> entrypoint::ProgramResult {
    let accounts_iter = &mut accounts.iter();
    let account = next_account_info(accounts_iter)?;

    msg!("Instruction data length: {}", instruction_data.len());
    msg!("Account data length: {}", account.data.borrow().len());

    // Deserialize the ZKP result from instruction data
    let zkp_result = ZkpResult::try_from_slice(instruction_data)
        .map_err(|_| ProgramError::InvalidInstructionData)?;
    
    msg!("Successfully deserialized ZKP result");
    msg!("Proof length: {}", zkp_result.proof.len());
    msg!("Verifying key length: {}", zkp_result.verifying_key.len());
    msg!("Public witness length: {}", zkp_result.public_witness.len());

    // Serialize to account data
    let mut account_data = account.data.borrow_mut();

    // Check if account has enough space
    let serialized_size = borsh::to_vec(&zkp_result)?.len();
    if account_data.len() < serialized_size {
        msg!("Account data size: {}, required size: {}", account_data.len(), serialized_size);
        return Err(ProgramError::AccountDataTooSmall);
    }
    
    zkp_result.serialize(&mut *account_data)
        .map_err(|_| ProgramError::AccountDataTooSmall)?;
    
    msg!("Successfully stored ZKP result in account");
    Ok(())
}
