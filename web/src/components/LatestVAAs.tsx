import { ChainId, tryHexToNativeString } from "@certusone/wormhole-sdk";
import { _parseVAAAlgorand } from "@certusone/wormhole-sdk/lib/esm/algorand/Algorand";
import { ArrowDownward, ArrowRight, ChevronRight } from "@mui/icons-material";
import { Card, IconButton, Typography } from "@mui/material";
import { Box } from "@mui/system";
import {
  createColumnHelper,
  getCoreRowModel,
  getExpandedRowModel,
  getSortedRowModel,
  Row,
  SortingState,
  useReactTable,
} from "@tanstack/react-table";
import { BigNumber } from "ethers";
import { ReactElement, useState } from "react";
import useLatestVAAs, { VAAsResponse } from "../hooks/useLatestVAAs";
import Table from "./Table";

const columnHelper = createColumnHelper<VAAsResponse>();

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
  columnHelper.accessor("_id", {
    id: "chain",
    header: () => "Chain",
    cell: (info) => info.getValue().split("/")[0],
  }),
  columnHelper.accessor("_id", {
    id: "emitter",
    header: () => "Emitter",
    cell: (info) => info.getValue().split("/")[1],
  }),
  columnHelper.accessor("_id", {
    id: "sequence",
    header: () => "Sequence",
    cell: (info) => info.getValue().split("/")[2],
  }),
  columnHelper.accessor("createdAt", {
    header: () => "Observed At",
    cell: (info) => new Date(info.getValue()).toLocaleString(),
  }),
];

function VAADetails({ row }: { row: Row<VAAsResponse> }): ReactElement {
  const parsedVaa = _parseVAAAlgorand(
    new Uint8Array(Buffer.from(row.original.vaa, "base64"))
  );
  let token = parsedVaa.Contract;
  // FromChain is a misnomer - actually OriginChain
  if (parsedVaa.Contract && parsedVaa.FromChain)
    try {
      token = tryHexToNativeString(
        parsedVaa.Contract,
        parsedVaa.FromChain as ChainId
      );
    } catch (e) {}
  return (
    <>
      Version: {parsedVaa.version}
      <br />
      Timestamp: {new Date(parsedVaa.timestamp * 1000).toLocaleString()}
      <br />
      Consistency: {parsedVaa.consistency}
      <br />
      Nonce: {parsedVaa.nonce}
      <br />
      Origin: {parsedVaa.FromChain}
      <br />
      Token: {token}
      <br />
      Amount: {BigNumber.from(parsedVaa.Amount).toString()}
      <br />
    </>
  );
}

function LatestVAAs() {
  const vaas = useLatestVAAs();
  const [sorting, setSorting] = useState<SortingState>([]);
  const table = useReactTable({
    columns,
    data: vaas,
    state: {
      sorting,
    },
    getRowId: (vaa) => vaa._id,
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
          <Typography variant="h5">Latest Messages</Typography>
        </Box>
        <Table<VAAsResponse> table={table} renderSubComponent={VAADetails} />
      </Card>
    </Box>
  );
}
export default LatestVAAs;
