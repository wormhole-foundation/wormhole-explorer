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
import useGovernorStatusChainDetails, {
  GovernorStatusChainDetailsResponse,
} from "../hooks/useGovernorStatusDetails";

import Table from "./Table";

const columnHelper = createColumnHelper<GovernorStatusChainDetailsResponse>();

const columns = [
  columnHelper.accessor("_id", {
    header: () => "Address",
    cell: (info) => (
      <Box component="pre" m={0}>
        {info.getValue()}
      </Box>
    ),
  }),
  columnHelper.accessor("createdAt", {
    header: () => "Created",
    cell: (info) => (info.getValue() ? info.getValue().toLocaleString() : null),
  }),
  columnHelper.accessor("updatedAt", {
    header: () => "LastUpdated",
    cell: (info) => (info.getValue() ? info.getValue().toLocaleString() : null),
  }),
  columnHelper.accessor("nodeName", {
    header: () => "Guardian",
    cell: (info) => (
      <Box component="pre" m={0}>
        {info.getValue()}
      </Box>
    ),
  }),
  columnHelper.accessor("chainId", {
    header: () => "Chain",
    cell: (info) => (
      <Box component="pre" m={0}>
        {info.getValue()}
      </Box>
    ),
  }),
  columnHelper.accessor("availableNotional", {
    header: () => "Available Notional (USD)",
    cell: (info) => (
      <Box textAlign="left">${numeral(info.getValue()).format("0,0")}</Box>
    ),
  }),
];

function GovStatusDetails(id: string) {
  const govStatusDetails = useGovernorStatusChainDetails(id);
  const [sorting, setSorting] = useState<SortingState>([]);
  const table = useReactTable({
    columns,
    data: govStatusDetails,
    state: {
      sorting,
    },
    getRowId: (chain) => chain._id,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: setSorting,
  });
  return (
    <Box m={2}>
      <Card>
        <Table<GovernorStatusChainDetailsResponse> table={table} />
      </Card>
    </Box>
  );
}

export default GovStatusDetails;
