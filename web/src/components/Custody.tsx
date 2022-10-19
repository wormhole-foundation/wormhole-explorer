import { ChevronRight } from "@mui/icons-material";
import { Box } from "@mui/system";
import { Card, IconButton, Typography } from "@mui/material";

import {
  createColumnHelper,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
  getExpandedRowModel,
  Row,
} from "@tanstack/react-table";
import numeral from "numeral";
import { useState, ReactElement } from "react";
import useCustodyData, { CustodyDataResponse } from "../hooks/useCustodyData";

import Table from "./Table";
import TokenDetails from "./TokenDetails";

const columnHelper = createColumnHelper<CustodyDataResponse>();

const columns = [
  columnHelper.display({
    id: "_expand",
    cell: ({ row }) =>
      row.getCanExpand() ? (
        <IconButton
          size="small"
          {...{
            onClick: row.getToggleExpandedHandler(),
            style: { cursor: "pointer" },
          }}
        >
          <ChevronRight
            sx={{
              transition: ".2s",
              transform: row.getIsExpanded() ? "rotate(90deg)" : undefined,
            }}
          />
        </IconButton>
      ) : null,
  }),
  columnHelper.accessor("chainId", {
    header: () => "Chain Id",
    sortingFn: `text`,
  }),
  columnHelper.accessor("chainName", {
    header: () => "Chain Name",
  }),
  columnHelper.accessor("custodyUSD", {
    header: () => "Total Value Locked (USD)",
    cell: (info) => (
      <Box textAlign="left">${numeral(info.getValue()).format("0,0.0000")}</Box>
    ),
  }),
  columnHelper.accessor("tokens", {
    header: () => "Locked Tokens",
    cell: (info) => {
      const value = info.getValue();
      return `${value.length} Token` + (value.length == 1 ? "" : `s`);
    },
  }),
];

/*
interface Token {
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

function AddTokenDetails({
  row,
}: {
  row: Row<CustodyDataResponse>;
}): ReactElement {
  const id = row.original._id;
  return TokenDetails(id);
}

function CustodyData() {
  const custody = useCustodyData();
  const [sorting, setSorting] = useState<SortingState>([]);
  const lockedValue = custody
    .map((x) => x.custodyUSD)
    .reduce((partialSum, a) => partialSum + a, 0);
  const table = useReactTable({
    columns,
    data: custody,
    state: {
      sorting,
    },
    getRowId: (chain) => chain.chainName,
    getRowCanExpand: () => true,
    getCoreRowModel: getCoreRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: setSorting,
  });
  return (
    <Box m={2}>
      <Card>
        <Box m={2}>
          Total Locked Value (USD): ${numeral(lockedValue).format("0,0.0000")}
        </Box>
        <Table<CustodyDataResponse>
          table={table}
          renderSubComponent={AddTokenDetails}
        />
      </Card>
    </Box>
  );
}

export default CustodyData;
