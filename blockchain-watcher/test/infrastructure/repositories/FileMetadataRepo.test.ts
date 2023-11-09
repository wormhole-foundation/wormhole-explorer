import { describe, expect, it, beforeEach, afterEach } from "@jest/globals";
import fs from "fs";
import { FileMetadataRepo } from "../../../src/infrastructure/repositories";

describe("FileMetadataRepo", () => {
  const dirPath = "./metadata-repo";
  const repo = new FileMetadataRepo(dirPath);

  beforeEach(() => {
    if (!fs.existsSync(dirPath)) {
      fs.mkdirSync(dirPath);
    }
  });

  afterEach(() => {
    fs.rm(dirPath, () => {});
  });

  describe("get", () => {
    it("should return null if the file does not exist", async () => {
      const metadata = await repo.get("non-existent-file");
      expect(metadata).toBeNull();
    });

    it("should return the metadata if the file exists", async () => {
      const id = "test-file";
      const metadata = { foo: "bar" };
      await repo.save(id, metadata);

      const retrievedMetadata = await repo.get(id);
      expect(retrievedMetadata).toEqual(metadata);
    });
  });

  describe("save", () => {
    it("should create a new file with the given metadata", async () => {
      const id = "test-file";
      const metadata = { foo: "bar" };
      await repo.save(id, metadata);

      const fileExists = fs.existsSync(`${dirPath}/${id}.json`);
      expect(fileExists).toBe(true);

      const fileContents = fs.readFileSync(`${dirPath}/${id}.json`, "utf8");
      expect(JSON.parse(fileContents)).toEqual(metadata);
    });

    it("should overwrite an existing file with the given metadata", async () => {
      const id = "test-file";
      const initialMetadata = { foo: "bar" };
      const updatedMetadata = { baz: "qux" };
      await repo.save(id, initialMetadata);
      await repo.save(id, updatedMetadata);

      const fileContents = fs.readFileSync(`${dirPath}/${id}.json`, "utf8");
      expect(JSON.parse(fileContents)).toEqual(updatedMetadata);
    });
  });
});
