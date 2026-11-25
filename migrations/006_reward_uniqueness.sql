CREATE UNIQUE INDEX contribution_reward_once
    ON credit_movements(operation_key)
    WHERE reason='contribution';
