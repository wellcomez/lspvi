function ts_a(): string {
    return ""
}
// 定义接口
interface PersonInterface {
    name: string;
    age: number;
    introduce(): void;
}

// 定义类并实现接口
class Person implements PersonInterface {
    name: string;
    age: number;

    constructor(name: string, age: number) {
        this.name = name;
        this.age = age;
    }

    introduce(): void {
        console.log(`Hello, my name is ${this.name} and I am ${this.age} years old.`);
    }
}
