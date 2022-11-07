import { Box, Card } from "@mui/material";
import {
  createColumnHelper,
  getCoreRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from "@tanstack/react-table";
import { useState } from "react";
import useEnqueuedVaaDetails, {
  EnqueuedVaaDetailsResponse,
} from "../hooks/useEnqueuedVaaDetails";
import Table from "./Table";
import numeral from "numeral";

// async function VaaExists(row: EnqueuedVaaDetailsResponse) {
//   try {
//     const chainId: ChainId = findChainId(row.chainId.toString());
//     const vaa = await getSignedVAAWithRetry(
//       WORMHOLE_RPC_HOSTS,
//       chainId,
//       row.emitterAddress,
//       row.sequence.toString(),
//       { transport: NodeHttpTransport() },
//       1000,
//       4
//     );
//     if (vaa != undefined) {
//       return true;
//     }
//   } catch (e) {
//     return false;
//   }
// }

const columnHelper = createColumnHelper<EnqueuedVaaDetailsResponse>();

const columns = [
  columnHelper.accessor("chainId", {
    header: () => "Chain Id",
    cell: (info) => (
      <Box component="pre" m={0}>
        {info.getValue()}
      </Box>
    ),
  }),
  columnHelper.accessor("emitterAddress", {
    header: () => "Emitter Address",
    cell: (info) => (
      <Box component="pre" m={0}>
        {info.getValue()}
      </Box>
    ),
  }),
  columnHelper.accessor("sequence", {
    header: () => "Sequence",
    cell: (info) => (
      <Box component="pre" m={0}>
        {info.getValue()}
      </Box>
    ),
  }),
  columnHelper.accessor("notionalValue", {
    header: () => "Notional Value (USD)",
    cell: (info) => (
      <Box textAlign="left">${numeral(info.getValue()).format("0,0.0")}</Box>
    ),
  }),
  columnHelper.accessor("releaseTime", {
    header: () => "Release Time",
    cell: (info) =>
      info.getValue()
        ? new Date(info.getValue() * 1000).toLocaleString()
        : null,
  }),
  // columnHelper.display({
  //   id: "hasQuorum",
  //   header: () => "Has Quorum?",
  //   cell: (info) => {
  //     const value = VaaExists(info.row.original);
  //     return value;
  //   },
  // }),
  columnHelper.accessor("txHash", {
    header: () => "Transaction",
    cell: (info) => (
      <Box component="pre" m={0}>
        {info.getValue()}
      </Box>
    ),
  }),
];

function EnqueuedVaaDetails(id: string) {
  const enqueuedVaaDetails = useEnqueuedVaaDetails(id);
  const [sorting, setSorting] = useState<SortingState>([]);
  const table = useReactTable({
    columns,
    data: enqueuedVaaDetails,
    state: {
      sorting,
    },
    getRowId: (chain) => chain.chainId,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: setSorting,
  });
  return (
    <Box m={2}>
      <Card>
        <Table<EnqueuedVaaDetailsResponse> table={table} />
      </Card>
    </Box>
  );
}

export default EnqueuedVaaDetails;
