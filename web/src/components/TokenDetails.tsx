import { Box, Card } from "@mui/material";
import {
  createColumnHelper,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from "@tanstack/react-table";
import numeral from "numeral";
import { useState } from "react";
import useTokenDetails, {
  TokenDetailsResponse,
} from "../hooks/useTokenDetails";

import Table from "./Table";
/*
export type TokenDetailsResponse {
  tokenAddress: string;
  name: string;
  decimals: number;
  symbol: string;
  balance: BigInt;
  qty: number;
  tokenPrice: number;
  tokenBalanceUSD: number;
}
*/
const columnHelper = createColumnHelper<TokenDetailsResponse>();

const columns = [
  columnHelper.accessor("tokenAddress", {
    header: () => "Token Address",
    sortingFn: `text`,
  }),
  columnHelper.accessor("name", {
    header: () => "Name",
  }),
  columnHelper.accessor("symbol", {
    header: () => "Symbol",
  }),
  columnHelper.accessor("qty", {
    header: () => "Token Balance",
    cell: (info) => (
      <Box textAlign="left">{numeral(info.getValue()).format("0,0.0000")}</Box>
    ),
  }),
  columnHelper.accessor("tokenPrice", {
    header: () => "Token Price",
    cell: (info) => (
      <Box textAlign="left">${numeral(info.getValue()).format("0,0.0000")}</Box>
    ),
  }),
  columnHelper.accessor("tokenBalanceUSD", {
    header: () => "Locked Value (USD)",
    cell: (info) => (
      <Box textAlign="left">${numeral(info.getValue()).format("0,0.0000")}</Box>
    ),
  }),
];

function TokenDetails(id: string) {
  const tokenDetails = useTokenDetails(id);
  const [sorting, setSorting] = useState<SortingState>([]);
  const table = useReactTable({
    columns,
    data: tokenDetails,
    state: {
      sorting,
    },
    getRowId: (token) => token.name,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: setSorting,
  });
  return (
    <Box m={2}>
      <Card>
        <Table<TokenDetailsResponse> table={table} />
      </Card>
    </Box>
  );
}

export default TokenDetails;
