import fs from "fs";
import { MetadataRepository } from "../../domain/repositories";

export class FileMetadataRepo implements MetadataRepository<any> {
  private readonly dirPath: string;

  constructor(dirPath: string) {
    this.dirPath = dirPath;
  }

  async get(id: string): Promise<any> {
    const filePath = `${this.dirPath}/${id}.json`;
    return fs.promises
      .readFile(filePath, "utf8")
      .then(JSON.parse)
      .catch((err) => null);
  }

  async save(id: string, metadata: any): Promise<void> {
    const filePath = `${this.dirPath}/${id}.json`;
    return fs.promises.writeFile(filePath, JSON.stringify(metadata), "utf8");
  }
}
