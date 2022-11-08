import { Box, IconButton, Card } from "@mui/material";
import { ChevronRight } from "@mui/icons-material";

import {
  createColumnHelper,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
  getExpandedRowModel,
  Row,
} from "@tanstack/react-table";
import { useState, ReactElement } from "react";
import useGovernorStatus, {
  GovernorStatusResponse,
} from "../hooks/useGovernorStatus";

import Table from "./Table";
import GovStatusDetails from "./GovStatusDetails";
import useEnqueuedVaas, {
  EnqueuedVaasResponse,
} from "../hooks/useEnqueuedVaas";
import EnqueuedVaaDetails from "./EnqueuedVaaDetails";
import numeral from "numeral";

const columnHelper = createColumnHelper<GovernorStatusResponse>();

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
    header: () => "ChainId",
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
  columnHelper.accessor("notionalLimit", {
    header: () => "Notional Limit (USD)",
    cell: (info) => (
      <Box textAlign="left">${numeral(info.getValue()).format("0,0")}</Box>
    ),
  }),
  columnHelper.accessor("maxTransactionSize", {
    header: () => "Max Transaction Size (USD)",
    cell: (info) => (
      <Box textAlign="left">${numeral(info.getValue()).format("0,0")}</Box>
    ),
  }),
];

const columnHelperVaa = createColumnHelper<EnqueuedVaasResponse>();

const columnsVaa = [
  columnHelperVaa.display({
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
  columnHelperVaa.accessor("chainId", {
    header: () => "ChainId",
    cell: (info) => (
      <Box component="pre" m={0}>
        {info.getValue()}
      </Box>
    ),
  }),
  columnHelperVaa.accessor("enqueuedVaas", {
    header: () => "Enqueued VAAs",
    cell: (info) => {
      const value = info.getValue();
      return `${value.length}`;
    },
  }),
];

function AddGovernorStatusDetails({
  row,
}: {
  row: Row<GovernorStatusResponse>;
}): ReactElement {
  const id = row.original.chainId.toString();
  return GovStatusDetails(id);
}

function AddEnqueuedVaas({
  row,
}: {
  row: Row<EnqueuedVaasResponse>;
}): ReactElement {
  const id = row.original.chainId.toString();
  return EnqueuedVaaDetails(id);
}

function GovernorStatus() {
  const governorStatus = useGovernorStatus();
  console.log(governorStatus);
  const [sorting, setSorting] = useState<SortingState>([]);

  const governorStatusTable = useReactTable({
    columns,
    data: governorStatus,
    state: {
      sorting,
    },
    getRowId: (governorStatus) => governorStatus.chainId.toString(),
    getRowCanExpand: () => true,
    getCoreRowModel: getCoreRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: setSorting,
  });

  const enqueuedVaas = useEnqueuedVaas();
  const enqueuedVAATable = useReactTable({
    columns: columnsVaa,
    data: enqueuedVaas,
    state: {
      sorting,
    },
    getRowId: (enqueuedVaa) => enqueuedVaa.chainId.toString(),
    getRowCanExpand: () => true,
    getCoreRowModel: getCoreRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: setSorting,
  });
  return (
    <Box m={2}>
      <Box m={2}>
        Available Notional By Chain:
        <Card>
          <Table<GovernorStatusResponse>
            table={governorStatusTable}
            renderSubComponent={AddGovernorStatusDetails}
          />
        </Card>
      </Box>
      <Box m={2}>
        Enqueued VAAs:
        <Card>
          <Table<EnqueuedVaasResponse>
            table={enqueuedVAATable}
            renderSubComponent={AddEnqueuedVaas}
          />
        </Card>
      </Box>
    </Box>
  );
}

export default GovernorStatus;
