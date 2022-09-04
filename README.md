# cic-keep-alive

A tiny daemon which periodically checks all vouchers health status and calls `changePeriod` (which also calls `applyDemurrageLimited(0)`) for any idle vocuher. This keeps voucher `transfer` gas usage acceptably low.

The voucher list source can be changed in the periodic queuer to any source. By default it uses the `cic-dw` tokens table.

